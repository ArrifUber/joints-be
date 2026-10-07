package repository

import (
	"context"

	"gorm.io/gorm"
	"joints-be/modules/context/entity"
)

type ContextRepository interface {
	// Create menyimpan context chunk baru ke database.
	Create(ctx context.Context, c *entity.ContextChunk) error

	// FindBySessionID mengembalikan semua context chunk untuk satu session, urut by created_at.
	FindBySessionID(ctx context.Context, sessionID string) ([]entity.ContextChunk, error)

	// FindLatestBySessionID mengembalikan context chunk terakhir untuk satu session.
	// Digunakan untuk membangun "current context" saat memproses transcript baru.
	FindLatestBySessionID(ctx context.Context, sessionID string) (*entity.ContextChunk, error)
}

type contextRepositoryImpl struct {
	db *gorm.DB
}

func NewContextRepository(db *gorm.DB) ContextRepository {
	return &contextRepositoryImpl{db: db}
}

func (r *contextRepositoryImpl) Create(ctx context.Context, c *entity.ContextChunk) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *contextRepositoryImpl) FindBySessionID(ctx context.Context, sessionID string) ([]entity.ContextChunk, error) {
	var chunks []entity.ContextChunk
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&chunks).Error
	return chunks, err
}

func (r *contextRepositoryImpl) FindLatestBySessionID(ctx context.Context, sessionID string) (*entity.ContextChunk, error) {
	var chunk entity.ContextChunk
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		First(&chunk).Error
	if err != nil {
		return nil, err
	}
	return &chunk, nil
}

