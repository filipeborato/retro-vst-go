package domain

import "time"

type Product struct {
    ProductID   uint      `gorm:"column:product_id;primaryKey;autoIncrement"`
    ProductName string    `gorm:"column:product_name;type:TEXT;not null"`
    Description string    `gorm:"column:description;type:TEXT"`
    Price       float64   `gorm:"column:price;type:DECIMAL(10,2);not null;default:0"`
    CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
    UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

func (Product) TableName() string {
    return "products"
}