package gorm

import (
	"errors"

	"gorm.io/gorm"
	"retro-vst-go/domain"
	"retro-vst-go/repository"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(transaction *domain.Transaction) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Fetch user to check and update balance
		var user domain.User
		if err := tx.First(&user, "user_id = ?", transaction.UserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrUserNotFound
			}
			return err
		}

		// 2. Check if user has sufficient balance
		if user.CurrentBalance < transaction.TransactionValue {
			return repository.ErrInsufficientBalance
		}

		// 3. Deduct transaction value
		user.CurrentBalance -= transaction.TransactionValue

		// 4. Save updated user
		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		// 5. Create transaction record
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *TransactionRepository) GetByID(id uint) (*domain.Transaction, error) {
	var transaction domain.Transaction
	if err := r.db.First(&transaction, "transaction_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepository) GetByUserID(userID uint) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	if err := r.db.Where("user_id = ?", userID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
