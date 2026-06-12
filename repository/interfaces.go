package repository

import (
	"errors"
	"retro-vst-go/domain"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrProductNotFound     = errors.New("product not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrSessionNotFound     = errors.New("session not found")
	ErrEmailAlreadyExists  = errors.New("email already exists")
)

type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id uint) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	GetByGoogleID(googleID string) (*domain.User, error)
	Update(user *domain.User) error
}

type ProductRepository interface {
	Create(product *domain.Product) error
	GetByID(id uint) (*domain.Product, error)
	GetAll() ([]domain.Product, error)
	Update(product *domain.Product) error
	Delete(id uint) error
}

type PaymentRepository interface {
	Create(payment *domain.Payment) error // Will handle database transaction to update user balance if approved
	ApproveByExternalID(tempSessionID string, finalPaymentID string) error // Confirms payment by external session ID and updates user balance
	GetByID(id uint) (*domain.Payment, error)
	GetByUserID(userID uint) ([]domain.Payment, error)
}

type TransactionRepository interface {
	Create(tx *domain.Transaction) error // Will handle database transaction to check/deduct balance
	GetByID(id uint) (*domain.Transaction, error)
	GetByUserID(userID uint) ([]domain.Transaction, error)
}

type SessionRepository interface {
	Create(session *domain.Session) error
	GetByID(id uint) (*domain.Session, error)
	GetByUserID(userID uint) ([]domain.Session, error)
	DeleteByUserID(userID uint) error
}

type PricingRepository interface {
	GetPluginCredits(pluginName string) int
	GetCreditCost(credits int, currency string) float64
}
