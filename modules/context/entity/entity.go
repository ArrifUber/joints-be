package entity

import (
	"encoding/json"
	"time"
)

type ContextChunk struct {
	ID        string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SessionID string          `gorm:"type:uuid;not null;index"`
	Sequence  int64           `gorm:"not null"` // Sequence dari transcript yang menghasilkan context ini
	Type      string          `gorm:"type:varchar(50);not null"`
	Data      json.RawMessage `gorm:"type:jsonb;not null"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime"`
}

func (ContextChunk) TableName() string {
	return "context_chunks"
}

