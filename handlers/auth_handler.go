// handlers/auth_handler.go

package handlers

import (
    "net/http"    
    "strings"   

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
	"github.com/golang-jwt/jwt/v4"
	"time"

    "retro-vst-go/config"
    "retro-vst-go/domain"
)

type SignupInput struct {
    Name     string `json:"name" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

func SignupHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input SignupInput
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Verificar se já existe usuário com esse email
        var existingUser domain.User
        if err := db.Where("email = ?", strings.ToLower(input.Email)).First(&existingUser).Error; err == nil {
            // Se não houve erro, significa que usuário existe
            c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
            return
        }

        // Hash da senha
        hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
            return
        }

        // Criar novo usuário
        user := domain.User{
            FullName:         input.Name,
            Email:        strings.ToLower(input.Email),
            PasswordHash: string(hash),
        }
        if err := db.Create(&user).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
            return
        }

        //Se quiser retornar um JWT imediatamente, gere aqui
        token, err := CreateJWT(user)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Signup successful",
            "token": token,
        })
    }
}

type LoginInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

func LoginHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input LoginInput
        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Buscar usuário pelo email
        var user domain.User
        if err := db.Where("email = ?", strings.ToLower(input.Email)).First(&user).Error; err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
            return
        }

        // Checar se a senha confere
        if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
            return
        }

        // Gerar JWT
        token, err := CreateJWT(user) // Implementar a geração do token
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
            return
        }

        // Criar um registro de sessão
        session := domain.Session{
            UserID:     user.UserID,
            AuthMethod: "password",
        }
        if err := db.Create(&session).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create session"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Login successful",
            "token":   token,
        })
    }
}

type CustomClaims struct {
    UserID uint `json:"user_id"`
    jwt.RegisteredClaims
}

func CreateJWT(user domain.User) (string, error) {
    config.LoadEnv()    
    claims := CustomClaims{
        UserID: user.UserID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            Issuer:    "retro-vst-go",
        },
    }  

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(config.JWTSecret))
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        config.LoadEnv()
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
            return
        }

        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

        token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
            return []byte(config.JWTSecret), nil
        })
        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }

        // Recupera claims
        claims, ok := token.Claims.(*CustomClaims)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
            return
        }

        // Salvar user_id nos context
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}