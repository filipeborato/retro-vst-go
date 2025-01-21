package domain

import "time"

type Payment struct {
    PaymentID          uint      `gorm:"column:payment_id;primaryKey;autoIncrement"`
    UserID             uint      `gorm:"column:user_id;not null"`
    ProductID          *uint     `gorm:"column:product_id"` // pode ser nulo
    ExternalPaymentID  string    `gorm:"column:external_payment_id;type:TEXT;not null"`
    SupplierName       string    `gorm:"column:supplier_name;type:TEXT;not null"`
    TopUpValue         float64   `gorm:"column:top_up_value;type:DECIMAL(10,2);not null;default:0"`
    BalanceAfterTopUp  float64   `gorm:"column:balance_after_top_up;type:DECIMAL(10,2);not null;default:0"`
    PaymentDate        time.Time `gorm:"column:payment_date;not null;default:CURRENT_TIMESTAMP"`
}

func (Payment) TableName() string {
    return "payments"
}
