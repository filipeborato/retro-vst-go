package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"retro-vst-go/domain"
	"retro-vst-go/repository"
)

type TransactionInput struct {
	ProductID uint `json:"product_id" binding:"required"`
}

func CreateTransactionHandler(
	transactionRepo repository.TransactionRepository,
	productRepo repository.ProductRepository,
) gin.HandlerFunc {
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

		var input TransactionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 1. Fetch product to verify and get price
		product, err := productRepo.GetByID(input.ProductID)
		if err != nil {
			if errors.Is(err, repository.ErrProductNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
			}
			return
		}

		// 2. Build transaction with current product price
		transaction := domain.Transaction{
			UserID:           userID,
			ProductID:        product.ProductID,
			TransactionValue: product.Price,
		}

		// 3. Create transaction (will debit user balance)
		if err := transactionRepo.Create(&transaction); err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			} else if errors.Is(err, repository.ErrInsufficientBalance) {
				c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient balance"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete transaction"})
			}
			return
		}

		c.JSON(http.StatusCreated, transaction)
	}
}

func GetUserTransactionsHandler(transactionRepo repository.TransactionRepository) gin.HandlerFunc {
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

		transactions, err := transactionRepo.GetByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
			return
		}

		c.JSON(http.StatusOK, transactions)
	}
}
