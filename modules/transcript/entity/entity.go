package entity

import "time"

// TranscriptStatus defines valid status values for a transcript chunk.
type TranscriptStatus string

const (
	TranscriptStatusPending    TranscriptStatus = "pending"
	TranscriptStatusProcessing TranscriptStatus = "processing"
	TranscriptStatusCompleted  TranscriptStatus = "completed"
	TranscriptStatusFailed     TranscriptStatus = "failed"
)

type TranscriptChunk struct {
	ID        string           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SessionID string           `gorm:"type:uuid;not null;index;uniqueIndex:idx_session_sequence"`
	Sequence  int64            `gorm:"not null;uniqueIndex:idx_session_sequence"`
	Text      string           `gorm:"type:text;not null"`
	Status    TranscriptStatus `gorm:"type:varchar(20);not null;default:'pending'"`
	CreatedAt time.Time        `gorm:"autoCreateTime"`
	UpdatedAt time.Time        `gorm:"autoUpdateTime"`
}

// TableName overrides default table name.
func (TranscriptChunk) TableName() string {
	return "transcript_chunks"
}
