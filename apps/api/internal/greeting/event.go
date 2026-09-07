package greeting

import (
	"encoding/json"
	"time"
)

const (
	StreamID            = "greeting"
	TypeGreetingSeeded  = "GreetingSeeded"
	TypeGreetingRenamed = "GreetingRenamed"
)

// Event is an immutable fact in the write-side log (Postgres).
type Event struct {
	StreamID   string          `json:"streamId"`
	Version    int64           `json:"version"`
	Type       string          `json:"type"`
	Payload    json.RawMessage `json:"payload"`
	OccurredAt time.Time       `json:"occurredAt"`
}

type GreetingPayload struct {
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updatedAt"`
	UpdatedBy string    `json:"updatedBy"`
}

func NewRenamed(version int64, rec Record, typ string) (Event, error) {
	if typ == "" {
		typ = TypeGreetingRenamed
	}
	payload, err := json.Marshal(GreetingPayload{
		Name:      rec.Name,
		UpdatedAt: rec.UpdatedAt.UTC(),
		UpdatedBy: rec.UpdatedBy,
	})
	if err != nil {
		return Event{}, err
	}
	return Event{
		StreamID:   StreamID,
		Version:    version,
		Type:       typ,
		Payload:    payload,
		OccurredAt: time.Now().UTC(),
	}, nil
}
