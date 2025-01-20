package main

import (
    "log"
    "os"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"

    "retro-vst-go/db"
)

func main() {
    // Carrega .env (opcional)
    if err := godotenv.Load(); err != nil {
        log.Printf("Não foi possível carregar .env: %v", err)
    }

    // Conecta ao banco, mas **não** chama AutoMigrate nem InsertMockData aqui
    database, err := db.SetupDatabase()
    if err != nil {
        log.Fatalf("Falha ao configurar o banco: %v\n", err)
    }

    //Se quiser fechar a conexão ao encerrar:
    defer func() {
       sqlDB, _ := database.DB()
       sqlDB.Close()
    }()

    r := gin.Default()

    // Rotas
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    port := os.Getenv("APP_PORT")
    if port == "" {
        port = ":8080"
    }
    r.Run(port)
}
