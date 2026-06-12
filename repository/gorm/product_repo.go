package gorm

import (
	"errors"

	"gorm.io/gorm"
	"retro-vst-go/domain"
	"retro-vst-go/repository"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *domain.Product) error {
	if err := r.db.Create(product).Error; err != nil {
		return err
	}
	return nil
}

func (r *ProductRepository) GetByID(id uint) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.First(&product, "product_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) GetAll() ([]domain.Product, error) {
	var products []domain.Product
	if err := r.db.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) Update(product *domain.Product) error {
	if err := r.db.Save(product).Error; err != nil {
		return err
	}
	return nil
}

func (r *ProductRepository) Delete(id uint) error {
	if err := r.db.Delete(&domain.Product{}, "product_id = ?", id).Error; err != nil {
		return err
	}
	return nil
}
