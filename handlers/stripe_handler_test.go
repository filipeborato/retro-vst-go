package handlers_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"retro-vst-go/config"
	"retro-vst-go/db"
	"retro-vst-go/domain"
	"retro-vst-go/handlers"
	gormrepo "retro-vst-go/repository/gorm"
)

// Helper para assinar o payload do webhook do Stripe simulando a assinatura real
func signStripeHeader(payload []byte, secret string, t time.Time) string {
	timestamp := strconv.FormatInt(t.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%s,v1=%s", timestamp, expectedSignature)
}

func TestEndToEndStripeAndPurchaseFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config.LoadEnv()

	stripeSecret := os.Getenv("STRIPE_SECRET_KEY")
	if stripeSecret == "" {
		t.Skip("Passe a chave STRIPE_SECRET_KEY no ambiente para rodar este teste")
		return
	}

	// 1. Configura as chaves reais de teste do Stripe fornecidas pelo ambiente para evitar expor segredos no git
	t.Setenv("STRIPE_SECRET_KEY", stripeSecret)
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test_secret_for_local_unit_testing")
	t.Setenv("JWT_KEY", "super-secret-test-jwt-key-12345678")
	t.Setenv("SQLITE_DB_PATH", filepath.Join(t.TempDir(), "test_flow.db"))

	config.LoadEnv()

	// 2. Setup do Banco e Repositórios
	dbConn, err := db.SetupDatabase()
	if err != nil {
		t.Fatalf("Erro ao configurar o banco de teste: %v", err)
	}
	if err := db.AutoMigrateDB(dbConn); err != nil {
		t.Fatalf("Erro ao migrar o banco: %v", err)
	}

	userRepo := gormrepo.NewUserRepository(dbConn)
	productRepo := gormrepo.NewProductRepository(dbConn)
	paymentRepo := gormrepo.NewPaymentRepository(dbConn)
	transactionRepo := gormrepo.NewTransactionRepository(dbConn)
	_ = gormrepo.NewSessionRepository(dbConn)

	// 3. Cadastra o Usuário
	user := domain.User{
		FullName:     "Test Buyer",
		Email:        "buyer@example.com",
		PasswordHash: "hashed-pw",
	}
	if err := userRepo.Create(&user); err != nil {
		t.Fatalf("Erro ao criar usuário: %v", err)
	}

	// 4. Cadastra o Produto
	product := domain.Product{
		ProductName: "PitchedDelay",
		Description: "Plugin de delay com pitch shifter",
		Price:       35.00,
	}
	if err := productRepo.Create(&product); err != nil {
		t.Fatalf("Erro ao criar produto: %v", err)
	}

	// 5. Gera JWT Token para o Usuário
	token, err := handlers.CreateJWT(user)
	if err != nil {
		t.Fatalf("Erro ao criar JWT: %v", err)
	}

	// 6. Setup do roteador HTTP
	r := gin.Default()
	r.POST("/payments/webhook/stripe", handlers.StripeWebhookHandler(paymentRepo))

	protected := r.Group("/api")
	protected.Use(handlers.AuthMiddleware())
	{
		protected.POST("/payments/stripe/checkout", handlers.CreateStripeCheckoutHandler(paymentRepo))
		protected.POST("/transactions", handlers.CreateTransactionHandler(transactionRepo, productRepo))
	}

	// ==========================================
	// PASSO A: Criar a Sessão de Checkout
	// ==========================================
	checkoutReqBody, _ := json.Marshal(map[string]interface{}{
		"amount": 100.00,
	})
	req, _ := http.NewRequest("POST", "/api/payments/stripe/checkout", bytes.NewBuffer(checkoutReqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Checkout falhou. Código de status: %d. Resposta: %s", w.Code, w.Body.String())
	}

	var checkoutRes struct {
		CheckoutURL string `json:"checkout_url"`
		SessionID   string `json:"session_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &checkoutRes); err != nil {
		t.Fatalf("Falha ao ler json de resposta do checkout: %v", err)
	}

	if checkoutRes.SessionID == "" || checkoutRes.CheckoutURL == "" {
		t.Fatalf("Stripe não retornou sessão válida: %+v", checkoutRes)
	}

	t.Logf("Sessão do Stripe criada com sucesso: %s", checkoutRes.SessionID)
	t.Logf("URL do Checkout: %s", checkoutRes.CheckoutURL)

	// Verifica se o pagamento pendente foi inserido no banco local
	var dbPayment domain.Payment
	err = dbConn.First(&dbPayment, "external_payment_id = ?", checkoutRes.SessionID).Error
	if err != nil {
		t.Fatalf("Erro ao buscar pagamento pendente no banco: %v", err)
	}
	if dbPayment.Status != "pending" || dbPayment.TopUpValue != 100.00 {
		t.Fatalf("Pagamento pendente incorreto no banco: %+v", dbPayment)
	}

	// ==========================================
	// PASSO B: Simular Confirmação do Stripe (Webhook)
	// ==========================================
	webhookPayload := []byte(fmt.Sprintf(`{
		"id": "evt_test_completed_123",
		"object": "event",
		"api_version": "2024-04-10",
		"type": "checkout.session.completed",
		"data": {
			"object": {
				"id": "%s",
				"object": "checkout.session",
				"payment_intent": "pi_test_transaction_success"
			}
		}
	}`, checkoutRes.SessionID))

	sig := signStripeHeader(webhookPayload, config.StripeWebhookSecret, time.Now())

	reqWebhook, _ := http.NewRequest("POST", "/payments/webhook/stripe", bytes.NewBuffer(webhookPayload))
	reqWebhook.Header.Set("Stripe-Signature", sig)
	reqWebhook.Header.Set("Content-Type", "application/json")

	wWebhook := httptest.NewRecorder()
	r.ServeHTTP(wWebhook, reqWebhook)

	if wWebhook.Code != http.StatusOK {
		t.Fatalf("Webhook falhou. Código de status: %d. Resposta: %s", wWebhook.Code, wWebhook.Body.String())
	}

	// Verifica se o saldo do usuário foi atualizado no banco
	var dbUser domain.User
	if err := dbConn.First(&dbUser, "user_id = ?", user.UserID).Error; err != nil {
		t.Fatalf("Erro ao buscar usuário: %v", err)
	}
	if dbUser.CurrentBalance != 100.00 {
		t.Fatalf("Saldo do usuário não foi creditado. Saldo atual: %.2f (esperado: 100.00)", dbUser.CurrentBalance)
	}

	// Verifica se o pagamento foi atualizado para "approved" e com o ID da transação final
	err = dbConn.First(&dbPayment, "external_payment_id = ?", "pi_test_transaction_success").Error
	if err != nil {
		t.Fatalf("Pagamento aprovado não encontrado pelo ID de transação final: %v", err)
	}
	if dbPayment.Status != "approved" || dbPayment.BalanceAfterTopUp != 100.00 {
		t.Fatalf("Dados do pagamento aprovado inconsistentes: %+v", dbPayment)
	}

	t.Logf("Saldo creditado com sucesso via Webhook! Saldo: R$%.2f", dbUser.CurrentBalance)

	// ==========================================
	// PASSO C: Efetuar a Compra do Produto (Dedução)
	// ==========================================
	transactionReqBody, _ := json.Marshal(map[string]interface{}{
		"product_id": product.ProductID,
	})
	reqTx, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(transactionReqBody))
	reqTx.Header.Set("Authorization", "Bearer "+token)
	reqTx.Header.Set("Content-Type", "application/json")

	wTx := httptest.NewRecorder()
	r.ServeHTTP(wTx, reqTx)

	if wTx.Code != http.StatusCreated {
		t.Fatalf("Compra falhou. Código de status: %d. Resposta: %s", wTx.Code, wTx.Body.String())
	}

	// Verifica se o saldo do usuário foi deduzido corretamente
	if err := dbConn.First(&dbUser, "user_id = ?", user.UserID).Error; err != nil {
		t.Fatalf("Erro ao buscar usuário: %v", err)
	}
	expectedBalance := 100.00 - 35.00
	if dbUser.CurrentBalance != expectedBalance {
		t.Fatalf("Saldo do usuário incorreto após a compra. Saldo: %.2f (esperado: %.2f)", dbUser.CurrentBalance, expectedBalance)
	}

	// Verifica se a transação foi salva no banco
	var dbTx domain.Transaction
	err = dbConn.First(&dbTx, "user_id = ? AND product_id = ?", user.UserID, product.ProductID).Error
	if err != nil {
		t.Fatalf("Transação não encontrada no banco: %v", err)
	}
	if dbTx.TransactionValue != 35.00 {
		t.Fatalf("Valor da transação incorreto: %.2f (esperado: 35.00)", dbTx.TransactionValue)
	}

	t.Logf("Compra efetuada com sucesso! Saldo final: R$%.2f", dbUser.CurrentBalance)
}

func TestStripeCheckoutLimitsAndProcessProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config.LoadEnv()

	stripeSecret := os.Getenv("STRIPE_SECRET_KEY")
	if stripeSecret == "" {
		t.Skip("Passe a chave STRIPE_SECRET_KEY no ambiente para rodar este teste")
		return
	}

	t.Setenv("STRIPE_SECRET_KEY", stripeSecret)
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test_secret_for_local_unit_testing")
	t.Setenv("JWT_KEY", "super-secret-test-jwt-key-12345678")
	t.Setenv("SQLITE_DB_PATH", filepath.Join(t.TempDir(), "test_limits.db"))

	config.LoadEnv()

	dbConn, err := db.SetupDatabase()
	if err != nil {
		t.Fatalf("Erro ao configurar o banco de teste: %v", err)
	}
	if err := db.AutoMigrateDB(dbConn); err != nil {
		t.Fatalf("Erro ao migrar o banco: %v", err)
	}

	userRepo := gormrepo.NewUserRepository(dbConn)
	paymentRepo := gormrepo.NewPaymentRepository(dbConn)

	// Cria usuário teste com saldo 0
	user := domain.User{
		FullName:       "Limit Tester",
		Email:          "limit@example.com",
		PasswordHash:   "hashed-pw",
		CurrentBalance: 0.00,
		Currency:       "BRL",
	}
	if err := userRepo.Create(&user); err != nil {
		t.Fatalf("Erro ao criar usuário: %v", err)
	}

	token, err := handlers.CreateJWT(user)
	if err != nil {
		t.Fatalf("Erro ao criar JWT: %v", err)
	}

	// Mock do servidor C++ VST Host
	vstServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("plugin") != "TheFunction" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mocked-processed-audio"))
	}))
	defer vstServer.Close()
	t.Setenv("VST_HOST_URL", vstServer.URL)

	r := gin.Default()
	protected := r.Group("/api")
	protected.Use(handlers.AuthMiddleware())
	{
		protected.POST("/payments/stripe/checkout", handlers.CreateStripeCheckoutHandler(paymentRepo))
		protected.POST("/process", handlers.CreateProcessProxyHandler(userRepo, dbConn))
	}

	// 1. Testa limite mínimo em BRL (deve falhar para R$ 4.99)
	reqBRL, _ := http.NewRequest("POST", "/api/payments/stripe/checkout", bytes.NewBufferString(`{"amount": 4.99, "currency": "BRL"}`))
	reqBRL.Header.Set("Authorization", "Bearer "+token)
	reqBRL.Header.Set("Content-Type", "application/json")
	wBRL := httptest.NewRecorder()
	r.ServeHTTP(wBRL, reqBRL)

	if wBRL.Code != http.StatusBadRequest {
		t.Fatalf("Deveria bloquear depósito menor que R$ 5.00. Código: %d, Resposta: %s", wBRL.Code, wBRL.Body.String())
	}
	t.Logf("Sucesso: bloqueado depósito de R$ 4.99 com erro: %s", wBRL.Body.String())

	// 2. Testa limite mínimo em USD (deve falhar para $ 0.99)
	reqUSD, _ := http.NewRequest("POST", "/api/payments/stripe/checkout", bytes.NewBufferString(`{"amount": 0.99, "currency": "USD"}`))
	reqUSD.Header.Set("Authorization", "Bearer "+token)
	reqUSD.Header.Set("Content-Type", "application/json")
	wUSD := httptest.NewRecorder()
	r.ServeHTTP(wUSD, reqUSD)

	if wUSD.Code != http.StatusBadRequest {
		t.Fatalf("Deveria bloquear depósito menor que $ 1.00. Código: %d, Resposta: %s", wUSD.Code, wUSD.Body.String())
	}
	t.Logf("Sucesso: bloqueado depósito de $ 0.99 com erro: %s", wUSD.Body.String())

	// 3. Testa processamento de áudio sem saldo (deve retornar 402)
	reqProc, _ := http.NewRequest("POST", "/api/process?plugin=TheFunction", bytes.NewBufferString("fake-audio-payload"))
	reqProc.Header.Set("Authorization", "Bearer "+token)
	reqProc.Header.Set("Content-Type", "audio/wav")
	wProc := httptest.NewRecorder()
	r.ServeHTTP(wProc, reqProc)

	if wProc.Code != http.StatusPaymentRequired {
		t.Fatalf("Deveria retornar 402 Payment Required para saldo insuficiente. Código: %d, Resposta: %s", wProc.Code, wProc.Body.String())
	}
	t.Logf("Sucesso: bloqueado processamento sem saldo com erro: %s", wProc.Body.String())

	// 4. Adiciona saldo manualmente para o usuário no banco (R$ 10.00)
	user.CurrentBalance = 10.00
	if err := userRepo.Update(&user); err != nil {
		t.Fatalf("Erro ao atualizar saldo do usuário: %v", err)
	}

	// 5. Roda o processamento com saldo (deve funcionar e debitar 10 créditos = R$ 0.50)
	wProcSuccess := httptest.NewRecorder()
	r.ServeHTTP(wProcSuccess, reqProc)

	if wProcSuccess.Code != http.StatusOK {
		t.Fatalf("Processamento com saldo deveria funcionar. Código: %d, Resposta: %s", wProcSuccess.Code, wProcSuccess.Body.String())
	}
	if wProcSuccess.Body.String() != "mocked-processed-audio" {
		t.Fatalf("Corpo de resposta de áudio incorreto: %s", wProcSuccess.Body.String())
	}

	// Verifica se o saldo foi deduzido em 10 créditos (10 * 0.05 = R$ 0.50) => saldo deve ser R$ 9.50
	var dbUser domain.User
	if err := dbConn.First(&dbUser, "user_id = ?", user.UserID).Error; err != nil {
		t.Fatalf("Erro ao buscar usuário: %v", err)
	}
	if dbUser.CurrentBalance != 9.50 {
		t.Fatalf("Saldo incorreto após processamento. Saldo: %.2f (esperado: 9.50)", dbUser.CurrentBalance)
	}

	t.Logf("Processamento efetuado com sucesso! Custo: 10 créditos (R$ 0.50). Saldo final: R$ %.2f", dbUser.CurrentBalance)
}

