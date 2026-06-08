package repository

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"retro-vst-go/domain"
)

type jsonPricingRepository struct {
	rules domain.PricingRules
}

func NewJSONPricingRepository(filePath string) PricingRepository {
	repo := &jsonPricingRepository{
		rules: domain.PricingRules{
			PluginCredits:         make(map[string]int),
			CreditConversionRates: make(map[string]float64),
		},
	}

	// Tenta achar o arquivo de regras subindo até 3 níveis de diretório (útil para testes em subpastas)
	resolvedPath := filePath
	for i := 0; i < 3; i++ {
		if _, err := os.Stat(resolvedPath); err == nil {
			break
		}
		resolvedPath = filepath.Join("..", resolvedPath)
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		log.Printf("Aviso: Não foi possível ler o arquivo de preços '%s' (%v). Usando padrões estáticos.", filePath, err)
		// Fallbacks estáticos caso falhe a leitura
		repo.rules.PluginCredits = map[string]int{
			"TheFunction":              10,
			"PitchedDelay":             8,
			"para-equalizer-x8-stereo": 6,
			"para-equalizer-x8-mono":   5,
			"compressor-stereo":        4,
			"filter-stereo":            2,
			"filter-mono":              1,
		}
		repo.rules.CreditConversionRates = map[string]float64{
			"USD": 0.01,
			"BRL": 0.05,
		}
		return repo
	}

	if err := json.Unmarshal(data, &repo.rules); err != nil {
		log.Printf("Aviso: Falha ao parsear JSON de preços: %v. Usando padrões.", err)
		repo.rules.PluginCredits = map[string]int{
			"TheFunction":              10,
			"PitchedDelay":             8,
			"para-equalizer-x8-stereo": 6,
			"para-equalizer-x8-mono":   5,
			"compressor-stereo":        4,
			"filter-stereo":            2,
			"filter-mono":              1,
		}
		repo.rules.CreditConversionRates = map[string]float64{
			"USD": 0.01,
			"BRL": 0.05,
		}
	}

	return repo
}

func (r *jsonPricingRepository) GetPluginCredits(pluginName string) int {
	if credits, exists := r.rules.PluginCredits[pluginName]; exists {
		return credits
	}
	return 3 // Custo padrão
}

func (r *jsonPricingRepository) GetCreditCost(credits int, currency string) float64 {
	rate := 0.05 // Default BRL
	if val, exists := r.rules.CreditConversionRates[currency]; exists {
		rate = val
	}
	return float64(credits) * rate
}
