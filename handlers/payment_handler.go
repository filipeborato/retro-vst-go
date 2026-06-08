package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"retro-vst-go/domain"
	"retro-vst-go/repository"
)

type PaymentInput struct {
	ExternalPaymentID string  `json:"external_payment_id" binding:"required"`
	SupplierName      string  `json:"supplier_name" binding:"required"`
	TopUpValue        float64 `json:"top_up_value" binding:"required,gt=0"`
}

func CreatePaymentHandler(paymentRepo repository.PaymentRepository) gin.HandlerFunc {
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

		var input PaymentInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		payment := domain.Payment{
			UserID:            userID,
			ExternalPaymentID: input.ExternalPaymentID,
			SupplierName:      input.SupplierName,
			TopUpValue:        input.TopUpValue,
		}

		if err := paymentRepo.Create(&payment); err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment top-up"})
			}
			return
		}

		c.JSON(http.StatusCreated, payment)
	}
}

func GetUserPaymentsHandler(paymentRepo repository.PaymentRepository) gin.HandlerFunc {
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

		payments, err := paymentRepo.GetByUserID(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
			return
		}

		c.JSON(http.StatusOK, payments)
	}
}
