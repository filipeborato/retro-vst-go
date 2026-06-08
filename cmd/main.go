package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"retro-vst-go/config"
	"retro-vst-go/db"
	"retro-vst-go/handlers"
	"retro-vst-go/repository"
	gormrepo "retro-vst-go/repository/gorm"
)

func main() {
	// Carrega as configurações de ambiente primeiro
	config.LoadEnv()
	if config.JWTSecret == "" {
		log.Fatal("ERRO CRÍTICO: A variável de ambiente JWT_KEY não está configurada.")
	}

	// Inicializa o banco de dados
	dbConn, err := db.SetupDatabase()
	if err != nil {
		log.Fatal(err)
	}

	// Inicializa os repositórios (Padrão Repository)
	userRepo := gormrepo.NewUserRepository(dbConn)
	productRepo := gormrepo.NewProductRepository(dbConn)
	paymentRepo := gormrepo.NewPaymentRepository(dbConn)
	transactionRepo := gormrepo.NewTransactionRepository(dbConn)
	sessionRepo := gormrepo.NewSessionRepository(dbConn)
	pricingRepo := repository.NewJSONPricingRepository("pricing_rules.json")

	// Configuração do CORS que suporta credenciais (cookies/headers de auth) de forma válida
	corsConfig := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true // Permite qualquer origem dinamicamente com credenciais habilitadas
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	r := gin.Default()
	r.Use(cors.New(corsConfig))

	// Rota de teste/ping
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Rotas de autenticação públicas
	r.POST("/signup", handlers.SignupHandler(userRepo))
	r.POST("/login", handlers.LoginHandler(userRepo, sessionRepo))

	// Google OAuth callback
	r.GET("/auth/google/callback", handlers.GoogleCallbackHandler(userRepo, sessionRepo))

	// Webhook do Stripe (Público, assinado)
	r.POST("/payments/webhook/stripe", handlers.StripeWebhookHandler(paymentRepo))

	// Catálogo público de produtos
	r.GET("/products", handlers.GetProductsHandler(productRepo))
	r.GET("/products/:id", handlers.GetProductByIDHandler(productRepo))

	// Rotas protegidas
	protected := r.Group("/api")
	protected.Use(handlers.AuthMiddleware())
	{
		// Perfil & Logout
		protected.GET("/profile", handlers.ProfileHandler(userRepo))
		protected.POST("/logout", handlers.LogoutHandler(sessionRepo))
		protected.POST("/process", handlers.CreateProcessProxyHandler(userRepo, pricingRepo, dbConn))

		// Administração de produtos
		protected.POST("/products", handlers.CreateProductHandler(productRepo))
		protected.PUT("/products/:id", handlers.UpdateProductHandler(productRepo))
		protected.DELETE("/products/:id", handlers.DeleteProductHandler(productRepo))

		// Carteira / Pagamentos (recargas)
		protected.POST("/payments", handlers.CreatePaymentHandler(paymentRepo))
		protected.GET("/payments", handlers.GetUserPaymentsHandler(paymentRepo))
		protected.POST("/payments/stripe/checkout", handlers.CreateStripeCheckoutHandler(paymentRepo))

		// Transações (compras)
		protected.POST("/transactions", handlers.CreateTransactionHandler(transactionRepo, productRepo))
		protected.GET("/transactions", handlers.GetUserTransactionsHandler(transactionRepo))
	}

	r.Run(":8080")
}
