package dto

import (
	"encoding/json"
	"time"
)

type ContextResponse struct {
	ID        string          `json:"id"`
	SessionID string          `json:"sessionId"`
	Sequence  int64           `json:"sequence"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"createdAt"`
}

