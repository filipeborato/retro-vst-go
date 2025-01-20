package domain

import "time"

type Transaction struct {
    TransactionID   uint      `gorm:"column:transaction_id;primaryKey;autoIncrement"`
    UserID          uint      `gorm:"column:user_id;not null"`
    ProductID       uint      `gorm:"column:product_id;not null"`
    TransactionValue float64  `gorm:"column:transaction_value;type:DECIMAL(10,2);not null;default:0"`
    TransactionDate time.Time `gorm:"column:transaction_date;not null;default:CURRENT_TIMESTAMP"`
}

func (Transaction) TableName() string {
    return "transactions"
}
