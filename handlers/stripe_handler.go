package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78"
	stripeSession "github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/webhook"
	"retro-vst-go/config"
	"retro-vst-go/domain"
	"retro-vst-go/repository"
)

type StripeCheckoutInput struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency"` // "BRL" ou "USD" (default: "BRL")
}

func CreateStripeCheckoutHandler(paymentRepo repository.PaymentRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}
		userID, ok := uid.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type in context"})
			return
		}

		var input StripeCheckoutInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		currency := strings.ToUpper(input.Currency)
		if currency == "" {
			currency = "BRL"
		}

		// Validação do limite mínimo de adição de saldo (R$ 5.00 ou $1.00)
		if currency == "BRL" && input.Amount < 5.00 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "O valor mínimo para recargas em Real é R$ 5,00"})
			return
		}
		if currency == "USD" && input.Amount < 1.00 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "The minimum top-up amount in USD is $1.00"})
			return
		}

		// Configura a chave secreta do Stripe
		stripe.Key = config.StripeSecretKey

		// 1. Cria a sessão de checkout no Stripe
		// NOTA: Omitimos "payment_method_types" para habilitar métodos de pagamento dinâmicos (cartão, pix, etc. configurados via painel do Stripe)
		params := &stripe.CheckoutSessionParams{
			Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency: stripe.String(strings.ToLower(currency)),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name:        stripe.String("RetroVST Wallet Top-up"),
							Description: stripe.String("Créditos para compra de plugins no RetroVST"),
						},
						UnitAmount: stripe.Int64(int64(input.Amount * 100)), // Em centavos
					},
					Quantity: stripe.Int64(1),
				},
			},
			SuccessURL: stripe.String(config.StripeSuccessURL + "?session_id={CHECKOUT_SESSION_ID}"),
			CancelURL:  stripe.String(config.StripeCancelURL),
			Metadata: map[string]string{
				"user_id": strconv.FormatUint(uint64(userID), 10),
			},
		}

		session, err := stripeSession.New(params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Stripe Checkout Session: " + err.Error()})
			return
		}

		// 2. Registra o pagamento local com status "pending" no banco local, vinculando o Session ID do Stripe
		payment := domain.Payment{
			UserID:            userID,
			ExternalPaymentID: session.ID,
			SupplierName:      "stripe",
			Status:            "pending",
			Currency:          currency,
			TopUpValue:        input.Amount,
		}

		if err := paymentRepo.Create(&payment); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save payment record"})
			return
		}

		// 3. Retorna a URL de redirecionamento para o checkout e o ID da sessão
		c.JSON(http.StatusOK, gin.H{
			"checkout_url": session.URL,
			"session_id":   session.ID,
		})
	}
}

func StripeWebhookHandler(paymentRepo repository.PaymentRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		const MaxBodyBytes = int64(65536)
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Verifica a assinatura do webhook do Stripe
		sigHeader := c.GetHeader("Stripe-Signature")
		event, err := webhook.ConstructEvent(payload, sigHeader, config.StripeWebhookSecret)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook signature: " + err.Error()})
			return
		}

		// Trata o evento de conclusão do checkout
		if event.Type == "checkout.session.completed" {
			var session stripe.CheckoutSession
			err := json.Unmarshal(event.Data.Raw, &session)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse webhook session details"})
				return
			}

			// O PaymentIntent representa a transação de cobrança final
			finalPaymentID := session.ID
			if session.PaymentIntent != nil {
				finalPaymentID = session.PaymentIntent.ID
			}

			if err := paymentRepo.ApproveByExternalID(session.ID, finalPaymentID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve payment: " + err.Error()})
				return
			}
		}

		c.Status(http.StatusOK)
	}
}
