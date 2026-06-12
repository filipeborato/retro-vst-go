package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"retro-vst-go/domain"
	"retro-vst-go/repository"
)

func getGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("URL_CALLBACK"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func GoogleCallbackHandler(userRepo repository.UserRepository, sessionRepo repository.SessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		config := getGoogleOAuthConfig()
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
			return
		}

		// Troca o code por token
		token, err := config.Exchange(context.Background(), code)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange token"})
			return
		}

		// Usa esse token para buscar info do usuário no endpoint do Google
		client := config.Client(context.Background(), token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user info"})
			return
		}
		defer resp.Body.Close()

		var googleUser struct {
			Id    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
			return
		}

		if googleUser.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email not provided by Google"})
			return
		}

		// 1. Verifica se já existe usuário com o GoogleID
		var user *domain.User
		user, err = userRepo.GetByGoogleID(googleUser.Id)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				// 2. Se não existir pelo GoogleID, tenta buscar pelo e-mail (vinculação automática)
				user, err = userRepo.GetByEmail(strings.ToLower(googleUser.Email))
				if err == nil {
					// Vincula o GoogleID ao usuário existente
					user.GoogleID = &googleUser.Id
					if err := userRepo.Update(user); err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link Google account"})
						return
					}
				} else if errors.Is(err, repository.ErrUserNotFound) {
					// 3. Se não existir nem pelo e-mail, cria um novo usuário
					newUser := domain.User{
						FullName:     googleUser.Name,
						Email:        strings.ToLower(googleUser.Email),
						GoogleID:     &googleUser.Id,
						PasswordHash: "oauth-google-placeholder", // Valor não nulo
					}
					if err := userRepo.Create(&newUser); err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user from Google account"})
						return
					}
					user = &newUser
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking email"})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking Google ID"})
				return
			}
		}

		// Gera JWT
		tokenString, err := CreateJWT(*user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Cria session com AuthMethod = "google"
		session := domain.Session{
			UserID:     user.UserID,
			AuthMethod: "google",
		}
		if err := sessionRepo.Create(&session); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":   tokenString,
			"message": "Login via Google successful",
		})
	}
}
