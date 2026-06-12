package domain

import "time"

// JSON Pricing models (used for JSON loading)
type PricingRules struct {
	PluginCredits         map[string]int     `json:"plugin_credits"`
	CreditConversionRates map[string]float64 `json:"credit_conversion_rates"`
}

// GORM/DB pricing models (prepared for future PostgreSQL/SQLite usage)
type PluginCredit struct {
	ID         uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PluginName string    `gorm:"column:plugin_name;type:varchar(255);uniqueIndex;not null" json:"plugin_name"`
	Credits    int       `gorm:"column:credits;not null" json:"credits"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (PluginCredit) TableName() string {
	return "plugin_credits"
}

type CreditRate struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Currency  string    `gorm:"column:currency;type:varchar(10);uniqueIndex;not null" json:"currency"`
	Rate      float64   `gorm:"column:rate;type:decimal(10,4);not null" json:"rate"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (CreditRate) TableName() string {
	return "credit_rates"
}
