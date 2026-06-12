package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"retro-vst-go/repository"
)

func ProfileHandler(userRepo repository.UserRepository) gin.HandlerFunc {
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

		user, err := userRepo.GetByID(userID)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":         user.UserID,
			"full_name":       user.FullName,
			"email":           user.Email,
			"current_balance": user.CurrentBalance,
		})
	}
}
