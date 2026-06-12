// handlers/auth_handler.go

package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"retro-vst-go/config"
	"retro-vst-go/domain"
	"retro-vst-go/repository"
)

type SignupInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func SignupHandler(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input SignupInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verificar se já existe usuário com esse email
		_, err := userRepo.GetByEmail(strings.ToLower(input.Email))
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
			return
		} else if !errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence"})
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
			FullName:     input.Name,
			Email:        strings.ToLower(input.Email),
			PasswordHash: string(hash),
		}
		if err := userRepo.Create(&user); err != nil {
			if errors.Is(err, repository.ErrEmailAlreadyExists) {
				c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
			}
			return
		}

		// Retornar um JWT imediatamente
		token, err := CreateJWT(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Signup successful",
			"token":   token,
		})
	}
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func LoginHandler(userRepo repository.UserRepository, sessionRepo repository.SessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Buscar usuário pelo email
		user, err := userRepo.GetByEmail(strings.ToLower(input.Email))
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user"})
			}
			return
		}

		// Checar se a senha confere
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Gerar JWT
		token, err := CreateJWT(*user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
			return
		}

		// Criar um registro de sessão
		session := domain.Session{
			UserID:     user.UserID,
			AuthMethod: "password",
		}
		if err := sessionRepo.Create(&session); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   token,
		})
	}
}

func LogoutHandler(sessionRepo repository.SessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}
		userID, ok := uid.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type in context"})
			return
		}

		if err := sessionRepo.DeleteByUserID(userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invalidate sessions"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
	}
}


type CustomClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func CreateJWT(user domain.User) (string, error) {
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