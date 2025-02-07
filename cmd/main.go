package main

import (    
    "log"
    "time" 
    
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "retro-vst-go/handlers"
    "retro-vst-go/db"        
)

func main() {
    // Conexão simples com SQLite (pode usar db.SetupDatabase() se tiver pronto)
    db, err := db.SetupDatabase()
    if err != nil {        
        log.Fatal(err)
    }
     // Configuração personalizada do CORS
     config := cors.Config{
        // Ajuste AllowOrigins conforme necessário; aqui estamos permitindo todas as origens
        AllowOrigins:     []string{"*"},
        // Permite os métodos HTTP necessários
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        // Permite os cabeçalhos, incluindo "Authorization"
        AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
        // Caso precise expor algum cabeçalho, configure aqui
        ExposeHeaders:    []string{"Content-Length"},
        // Se sua aplicação usa credenciais (cookies, etc.)
        AllowCredentials: true,
        // Define o tempo máximo de cache para o preflight
        MaxAge:           12 * time.Hour,
    }

    r := gin.Default()
    r.Use(cors.New(config))

    // Rotas signup e login
    r.POST("/signup", handlers.SignupHandler(db))
    r.POST("/login", handlers.LoginHandler(db))

    // Google OAuth
    // - Rota de redirect: GET /auth/google => redireciona para Google   
    r.GET("/auth/google/callback", handlers.GoogleCallbackHandler(db))

    // Exemplo de rota protegida
    // r.GET("/profile", AuthMiddleware(), ProfileHandler)

    protected := r.Group("/api")
    protected.Use(handlers.AuthMiddleware())
    {
        protected.GET("/profile", handlers.ProfileHandler(db))        
    }

    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    r.Run(":8080")
}

