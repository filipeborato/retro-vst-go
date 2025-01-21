package domain

import "time"

type User struct {
    UserID         uint    `gorm:"column:user_id;primaryKey;autoIncrement"`
    FullName       string  `gorm:"column:full_name;type:TEXT;not null"`
    Email          string  `gorm:"column:email;type:TEXT;unique;not null"`
    PasswordHash   string  `gorm:"column:password_hash;type:TEXT;not null"`
    CurrentBalance float64 `gorm:"column:current_balance;type:DECIMAL(10,2);not null;default:0"`
    CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
    UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

// TableName permite definir explicitamente o nome da tabela.
// Caso contrário, o GORM usaria "users".
func (User) TableName() string {
    return "users"
}
