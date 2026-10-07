package session

import (
	"time"
)

type Session struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Subject   string    `gorm:"type:varchar(100);not null"`
	Topic     string    `gorm:"type:varchar(100);not null"`
	Subtopic  string    `gorm:"type:varchar(100);not null"`
	Status    string    `gorm:"type:varchar(20);not null;default:'active'"` // active, completed, cancelled
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

