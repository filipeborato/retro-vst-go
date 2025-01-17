// main.go

package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"

    // Ajuste o nome do import conforme o seu módulo
    "retrovst/db"
)

func main() {	
    // 1. Carregamos as variáveis de ambiente do arquivo .env
    err := godotenv.Load()
    if err != nil {
        log.Printf("Não foi possível carregar .env: %v", err)
    }

    // 2. Se quisermos ler a porta do arquivo .env também
    port := os.Getenv("APP_PORT")
    if port == "" {
        port = ":8080"
    }

    // 3. Configura o banco de dados (SQLite) usando a função SetupDatabase()
    database, err := db.SetupDatabase()
    if err != nil {
        log.Fatalf("Falha ao configurar o banco: %v\n", err)
    }
    defer database.Close()

    // 4. Inicializa o Gin
    r := gin.Default()

    // Rota simples para teste
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    // Rota para testar a conexão com o banco
    r.GET("/testdb", func(c *gin.Context) {
        _, err := database.Exec("SELECT 1")
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        c.JSON(200, gin.H{"message": "DB connected successfully"})
    })

    // 5. Inicia o servidor na porta definida em .env
    r.Run(port)
}
