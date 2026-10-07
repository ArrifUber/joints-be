package dto

import "time"

type CreateSessionRequest struct {
	Subject  string `json:"subject" validate:"required"`
	Topic    string `json:"topic" validate:"required"`
	Subtopic string `json:"subtopic" validate:"required"`
}

type SessionResponse struct {
	ID        string    `json:"id"`
	Subject   string    `json:"subject"`
	Topic     string    `json:"topic"`
	Subtopic  string    `json:"subtopic"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

