package event

import (
	"context"
	"time"
)

// Event merepresentasikan payload event yang dikirimkan ke client (misal via WebSocket).
type Event struct {
	Event     string    `json:"event"`
	SessionID string    `json:"sessionId"`
	Sequence  int64     `json:"sequence"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

// EventPublisher adalah abstraksi untuk mempublikasikan event tanpa coupling ke WebSocket.
type EventPublisher interface {
	Publish(ctx context.Context, sessionID string, event Event) error
}

// NoopEventPublisher adalah implementasi default sebelum WebSocket hub diintegrasikan.
type NoopEventPublisher struct{}

func NewNoopEventPublisher() EventPublisher {
	return &NoopEventPublisher{}
}

func (n *NoopEventPublisher) Publish(ctx context.Context, sessionID string, event Event) error {
	return nil
}

