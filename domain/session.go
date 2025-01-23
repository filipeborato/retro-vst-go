package domain

import "time"

type Session struct {
    ID             uint      `gorm:"primaryKey;autoIncrement"`
    UserID         uint      `gorm:"not null"`
    AuthMethod     string    `gorm:"type:TEXT;not null"` // "password" ou "google"
    CreatedAt      time.Time
    // Caso queira guardar o token JWT ou refresh_token, adicione campos
    // Token        string
    // ExpiresAt    time.Time
}

func (Session) TableName() string {
    return "sessions"
}
