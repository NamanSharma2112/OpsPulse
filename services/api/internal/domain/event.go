package domain

import (
	"encoding/json"
	"time"
)

// Event is a raw GitHub webhook delivery, stored verbatim so projections can
// always be rebuilt from the source of truth.
type Event struct {
	ID         string          `json:"id"`
	ProjectID  string          `json:"project_id"`
	DeliveryID string          `json:"delivery_id"`
	Type       string          `json:"type"`
	Action     string          `json:"action,omitempty"`
	Actor      string          `json:"actor,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	OccurredAt time.Time       `json:"occurred_at"`
	ReceivedAt time.Time       `json:"received_at"`
}
