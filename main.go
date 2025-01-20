package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"

    "retro-vst-go/db" // Ajuste conforme seu go.mod
)

func main() {
    // Carrega .env se desejar
    if err := godotenv.Load(); err != nil {
        log.Printf("Não foi possível carregar .env: %v", err)
    }

    // Inicia o banco (SQLite)
    database, err := db.SetupDatabase()
    if err != nil {
        log.Fatalf("Falha ao configurar o banco: %v\n", err)
    }

    // Faz migração (AutoMigrate + índices + triggers)
    if err := db.AutoMigrateDB(database); err != nil {
        log.Fatalf("Erro na migração: %v\n", err)
    }

    // Se estiver em ambiente de teste/desenvolvimento, pode inserir mocks:
    // env := os.Getenv("APP_ENV") // ex: "development" ou "production"
    // if env == "development" {
    //     if err := db.InsertMockData(database); err != nil {
    //         log.Fatalf("Erro ao inserir dados de teste: %v\n", err)
    //     }
    // }

    // Inicializa Gin
    r := gin.Default()

    // Exemplo de rota teste
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    r.Run(":8080") // inicia servidor
}
