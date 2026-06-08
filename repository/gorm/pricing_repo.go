package gorm

import (
	"retro-vst-go/domain"
	"retro-vst-go/repository"
	"gorm.io/gorm"
)

type gormPricingRepository struct {
	db *gorm.DB
}

func NewGormPricingRepository(db *gorm.DB) repository.PricingRepository {
	return &gormPricingRepository{db: db}
}

func (r *gormPricingRepository) GetPluginCredits(pluginName string) int {
	var pc domain.PluginCredit
	if err := r.db.Where("plugin_name = ?", pluginName).First(&pc).Error; err == nil {
		return pc.Credits
	}
	return 3 // Custo padrão se não encontrar no banco
}

func (r *gormPricingRepository) GetCreditCost(credits int, currency string) float64 {
	var cr domain.CreditRate
	if err := r.db.Where("currency = ?", currency).First(&cr).Error; err == nil {
		return float64(credits) * cr.Rate
	}
	// Fallbacks estáticos padrão se não encontrar no banco
	rate := 0.05
	if currency == "USD" {
		rate = 0.01
	}
	return float64(credits) * rate
}
