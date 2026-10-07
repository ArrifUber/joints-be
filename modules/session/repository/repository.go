package repository

import (
	"context"

	"gorm.io/gorm"
	"joints-be/modules/session"
)

type SessionRepository interface {
	Create(ctx context.Context, s *session.Session) error
	GetByID(ctx context.Context, id string) (*session.Session, error)
}

type sessionRepositoryImpl struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepositoryImpl{db: db}
}

func (r *sessionRepositoryImpl) Create(ctx context.Context, s *session.Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *sessionRepositoryImpl) GetByID(ctx context.Context, id string) (*session.Session, error) {
	var s session.Session
	if err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

