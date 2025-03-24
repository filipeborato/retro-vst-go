package config

import (
    "log"
    "os"
    "github.com/joho/godotenv"
)

var JWTSecret string
var SQLiteDBPath string

func LoadEnv() {
    // Tenta carregar o .env (se não estiver em produção)
    err := godotenv.Load()
    if err != nil {
        log.Println("Aviso: Nenhum arquivo .env encontrado, usando variáveis de ambiente do sistema")
    }

    // Carregar a variável JWT_KEY
    JWTSecret = os.Getenv("JWT_KEY")    
	SQLiteDBPath = os.Getenv("SQLITE_DB_PATH")
}
