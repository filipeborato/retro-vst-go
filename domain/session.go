package domain

import "time"

type Session struct {
    SessionID  uint   `gorm:"column:session_id;primaryKey;autoIncrement"`
    UserID     uint   `gorm:"column:user_id;not null"`

    // Relacionamento com User (FK: user_id -> users.user_id)
    User *User `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

    AuthMethod string    `gorm:"type:TEXT;not null"` // "password" ou "google"
    CreatedAt  time.Time
    // Campos extras (token, expiresAt etc.) se precisar
}

// Nome da tabela
func (Session) TableName() string {
    return "sessions"
}
