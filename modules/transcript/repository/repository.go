package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"joints-be/modules/transcript/entity"
)

// ErrDuplicateSequence is returned when a transcript with the same
// session_id and sequence already exists.
var ErrDuplicateSequence = errors.New("transcript sequence already exists for this session")

type TranscriptRepository interface {
	// Create saves a new transcript chunk. Returns ErrDuplicateSequence on conflict.
	Create(ctx context.Context, t *entity.TranscriptChunk) error

	// FindPending returns up to `limit` pending transcript chunks ordered by sequence.
	FindPending(ctx context.Context, limit int) ([]entity.TranscriptChunk, error)

	// FindBySessionID returns all transcript chunks for a session ordered by sequence.
	FindBySessionID(ctx context.Context, sessionID string) ([]entity.TranscriptChunk, error)

	// UpdateStatus updates the status of a transcript chunk by ID.
	UpdateStatus(ctx context.Context, id string, status entity.TranscriptStatus) error
}

type transcriptRepositoryImpl struct {
	db *gorm.DB
}

func NewTranscriptRepository(db *gorm.DB) TranscriptRepository {
	return &transcriptRepositoryImpl{db: db}
}

func (r *transcriptRepositoryImpl) Create(ctx context.Context, t *entity.TranscriptChunk) error {
	result := r.db.WithContext(ctx).Create(t)
	if result.Error != nil {
		// Detect unique constraint violation on (session_id, sequence)
		if isDuplicateError(result.Error) {
			return ErrDuplicateSequence
		}
		return result.Error
	}
	return nil
}

func (r *transcriptRepositoryImpl) FindPending(ctx context.Context, limit int) ([]entity.TranscriptChunk, error) {
	var chunks []entity.TranscriptChunk
	err := r.db.WithContext(ctx).
		Where("status = ?", entity.TranscriptStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&chunks).Error
	return chunks, err
}

func (r *transcriptRepositoryImpl) FindBySessionID(ctx context.Context, sessionID string) ([]entity.TranscriptChunk, error) {
	var chunks []entity.TranscriptChunk
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("sequence ASC").
		Find(&chunks).Error
	return chunks, err
}

func (r *transcriptRepositoryImpl) UpdateStatus(ctx context.Context, id string, status entity.TranscriptStatus) error {
	return r.db.WithContext(ctx).
		Model(&entity.TranscriptChunk{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// isDuplicateError checks if the error is a PostgreSQL unique constraint violation.
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), "23505") || contains(err.Error(), "duplicate key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

