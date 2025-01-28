package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "retro-vst-go/domain"
)

func ProfileHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {        
        uid, exists := c.Get("user_id")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
            return
        }

        userID := uid.(uint)
        var user domain.User
        err := db.Where("user_id = ?", userID).First(&user).Error
        if err != nil {
            if err == gorm.ErrRecordNotFound {
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
            // Adicione mais campos se houver
        })
    }
}
