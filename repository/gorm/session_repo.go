package gorm

import (
	"errors"

	"gorm.io/gorm"
	"retro-vst-go/domain"
	"retro-vst-go/repository"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session *domain.Session) error {
	if err := r.db.Create(session).Error; err != nil {
		return err
	}
	return nil
}

func (r *SessionRepository) GetByID(id uint) (*domain.Session, error) {
	var session domain.Session
	if err := r.db.First(&session, "session_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrSessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) GetByUserID(userID uint) ([]domain.Session, error) {
	var sessions []domain.Session
	if err := r.db.Where("user_id = ?", userID).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *SessionRepository) DeleteByUserID(userID uint) error {
	if err := r.db.Delete(&domain.Session{}, "user_id = ?", userID).Error; err != nil {
		return err
	}
	return nil
}
