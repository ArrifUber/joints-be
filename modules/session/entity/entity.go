package entity

import (
	"time"
)

// SessionStatus defines valid status values for a session.
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusCancelled SessionStatus = "cancelled"
)

type Session struct {
	ID        string        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Subject   string        `gorm:"type:varchar(100);not null"`
	Topic     string        `gorm:"type:varchar(100);not null"`
	Subtopic  string        `gorm:"type:varchar(100);not null"`
	Status    SessionStatus `gorm:"type:varchar(20);not null;default:'active'"`
	CreatedAt time.Time     `gorm:"autoCreateTime"`
	UpdatedAt time.Time     `gorm:"autoUpdateTime"`
}

