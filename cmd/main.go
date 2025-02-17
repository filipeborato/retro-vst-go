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
    db, err := db.SetupDatabase()
    if err != nil {        
        log.Fatal(err)
    }
     // Configuração personalizada do CORS
     config := cors.Config{        
        AllowOrigins:     []string{"*"},        
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},        
        AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},        
        ExposeHeaders:    []string{"Content-Length"},      
        AllowCredentials: true,        
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

