package websocket

import (
	"context"

	"joints-be/pkg/event"
)

type HubEventPublisher struct {
	hub Hub
}

func NewHubEventPublisher(hub Hub) event.EventPublisher {
	return &HubEventPublisher{hub: hub}
}

func (p *HubEventPublisher) Publish(ctx context.Context, sessionID string, ev event.Event) error {
	p.hub.Broadcast(sessionID, ev)
	return nil
}

