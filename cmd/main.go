package main

import (    
    "log" 
    
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "retro-vst-go/handlers"
    "retro-vst-go/db"        
)

func main() {
    // Conexão simples com SQLite (pode usar db.SetupDatabase() se tiver pronto)
    db, err := db.SetupDatabase()
    if err != nil {        
        log.Fatal(err)
    }

    r := gin.Default()
    r.Use(cors.Default())

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

