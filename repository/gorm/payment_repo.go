package gorm

import (
	"errors"

	"gorm.io/gorm"
	"retro-vst-go/domain"
	"retro-vst-go/repository"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(payment *domain.Payment) error {
	// Se for pendente, apenas cria o registro sem atualizar saldo
	if payment.Status == "pending" {
		if err := r.db.Create(payment).Error; err != nil {
			return err
		}
		return nil
	}

	// Se for aprovado (ex: seeds/mock data), atualiza o saldo do usuário no ato da criação
	payment.Status = "approved"
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user domain.User
		if err := tx.First(&user, "user_id = ?", payment.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrUserNotFound
			}
			return err
		}

		user.CurrentBalance += payment.TopUpValue
		payment.BalanceAfterTopUp = user.CurrentBalance

		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *PaymentRepository) ApproveByExternalID(tempSessionID string, finalPaymentID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var payment domain.Payment
		if err := tx.First(&payment, "external_payment_id = ?", tempSessionID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("payment not found")
			}
			return err
		}

		// Idempotency check: se já estiver aprovado, não faz nada
		if payment.Status == "approved" {
			return nil
		}

		var user domain.User
		if err := tx.First(&user, "user_id = ?", payment.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrUserNotFound
			}
			return err
		}

		user.CurrentBalance += payment.TopUpValue
		payment.BalanceAfterTopUp = user.CurrentBalance
		payment.Status = "approved"
		payment.ExternalPaymentID = finalPaymentID // Guarda o ID final da transação

		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		if err := tx.Save(&payment).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *PaymentRepository) GetByID(id uint) (*domain.Payment, error) {
	var payment domain.Payment
	if err := r.db.First(&payment, "payment_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) GetByUserID(userID uint) ([]domain.Payment, error) {
	var payments []domain.Payment
	if err := r.db.Where("user_id = ?", userID).Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}
