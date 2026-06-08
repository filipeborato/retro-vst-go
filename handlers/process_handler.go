package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"retro-vst-go/domain"
	"retro-vst-go/repository"
)

// Mapeamento de créditos consumidos por plugin
func getPluginCredits(pluginName string) int {
	switch pluginName {
	case "TheFunction":
		return 10
	case "PitchedDelay":
		return 8
	case "para-equalizer-x8-stereo":
		return 6
	case "para-equalizer-x8-mono":
		return 5
	case "compressor-stereo":
		return 4
	case "filter-stereo":
		return 2
	case "filter-mono":
		return 1
	default:
		return 3 // Custo padrão para outros plugins
	}
}

// Conversão de créditos para moeda real
func getCreditCost(credits int, currency string) float64 {
	if currency == "USD" {
		return float64(credits) * 0.01 // 1 crédito = USD $0.01
	}
	return float64(credits) * 0.05 // 1 crédito = BRL R$ 0.05 (Default)
}

func CreateProcessProxyHandler(userRepo repository.UserRepository, dbConn *gorm.DB) gin.HandlerFunc {
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

		// 1. Identifica o plugin solicitado
		pluginName := c.Query("plugin")
		if pluginName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'plugin' is required"})
			return
		}

		creditsRequired := getPluginCredits(pluginName)

		// 2. Verifica saldo do usuário
		var cost float64
		var user *domain.User

		// Executa validação e débito em transação segura
		err := dbConn.Transaction(func(tx *gorm.DB) error {
			// Busca usuário atualizado
			var u domain.User
			if err := tx.First(&u, "user_id = ?", userID).Error; err != nil {
				return err
			}
			user = &u

			// Calcula o custo com base na moeda da conta do usuário
			cost = getCreditCost(creditsRequired, u.Currency)

			// Verifica se tem saldo suficiente
			if u.CurrentBalance < cost {
				return repository.ErrInsufficientBalance
			}

			return nil
		})

		if err != nil {
			if errors.Is(err, repository.ErrInsufficientBalance) {
				if user.Currency == "USD" {
					c.JSON(http.StatusPaymentRequired, gin.H{
						"error": fmt.Sprintf("Insufficient balance. This process requires %d credits ($%.2f). Your current balance is $%.2f.",
							creditsRequired, cost, user.CurrentBalance),
					})
				} else {
					c.JSON(http.StatusPaymentRequired, gin.H{
						"error": fmt.Sprintf("Saldo insuficiente. Este processamento requer %d créditos (R$ %.2f). Seu saldo atual é R$ %.2f.",
							creditsRequired, cost, user.CurrentBalance),
					})
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
			}
			return
		}

		// 3. Encaminha a requisição de processamento ao Host C++ VST local
		vstHostURL := os.Getenv("VST_HOST_URL")
		if vstHostURL == "" {
			vstHostURL = "http://localhost:18080"
		}
		cppHostURL := vstHostURL + "/process"
		if c.Request.URL.RawQuery != "" {
			cppHostURL += "?" + c.Request.URL.RawQuery
		}

		proxyReq, err := http.NewRequest("POST", cppHostURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request for VST host"})
			return
		}

		// Repassa cabeçalhos de multipart form-data
		proxyReq.Header.Set("Content-Type", c.GetHeader("Content-Type"))
		proxyReq.Header.Set("Content-Length", c.GetHeader("Content-Length"))

		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to VST Host server: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		// Se o processamento no host falhou, repassa o erro e não cobra saldo
		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"error": "VST Host processing failed"})
			return
		}

		// 4. Se o processamento deu certo, debita o saldo de forma segura
		err = dbConn.Transaction(func(tx *gorm.DB) error {
			var u domain.User
			if err := tx.First(&u, "user_id = ?", userID).Error; err != nil {
				return err
			}

			// Deduz o valor
			u.CurrentBalance -= cost
			return tx.Save(&u).Error
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to debit balance: " + err.Error()})
			return
		}

		// 5. Retorna o arquivo de áudio processado diretamente ao cliente
		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	}
}
