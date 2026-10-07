package repository

import (
	"context"

	"gorm.io/gorm"
	"joints-be/modules/session/entity"
)

type SessionRepository interface {
	Create(ctx context.Context, s *entity.Session) error
	GetByID(ctx context.Context, id string) (*entity.Session, error)
}

type sessionRepositoryImpl struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepositoryImpl{db: db}
}

func (r *sessionRepositoryImpl) Create(ctx context.Context, s *entity.Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *sessionRepositoryImpl) GetByID(ctx context.Context, id string) (*entity.Session, error) {
	var s entity.Session
	if err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}
