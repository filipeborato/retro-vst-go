package config

import (
    "log"
    "os"
    "github.com/joho/godotenv"
)

var JWTSecret string
var SQLiteDBPath string
var StripeSecretKey string
var StripeWebhookSecret string
var StripeSuccessURL string
var StripeCancelURL string

func LoadEnv() {
	// Tenta carregar o .env do diretório atual, se não achar tenta no diretório pai (útil para testes em subpastas)
	err := godotenv.Load()
	if err != nil {
		err = godotenv.Load("../.env")
		if err != nil {
			log.Println("Aviso: Nenhum arquivo .env encontrado, usando variáveis de ambiente do sistema")
		}
	}

	// Carregar a variável JWT_KEY
	JWTSecret = os.Getenv("JWT_KEY")
	SQLiteDBPath = os.Getenv("SQLITE_DB_PATH")

	// Configurações do Stripe
	StripeSecretKey = os.Getenv("STRIPE_SECRET_KEY")
	StripeWebhookSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")

	StripeSuccessURL = os.Getenv("STRIPE_SUCCESS_URL")
	if StripeSuccessURL == "" {
		StripeSuccessURL = "http://localhost:3000/payment/success"
	}

	StripeCancelURL = os.Getenv("STRIPE_CANCEL_URL")
	if StripeCancelURL == "" {
		StripeCancelURL = "http://localhost:3000/payment/cancel"
	}
}
