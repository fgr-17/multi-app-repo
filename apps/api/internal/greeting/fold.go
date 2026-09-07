package greeting

import (
	"encoding/json"
	"fmt"
)

// Fold rebuilds the current greeting from the event stream.
func Fold(events []Event) (Record, error) {
	var rec Record
	for _, e := range events {
		switch e.Type {
		case TypeGreetingSeeded, TypeGreetingRenamed:
			var p GreetingPayload
			if err := json.Unmarshal(e.Payload, &p); err != nil {
				return Record{}, fmt.Errorf("payload %s v%d: %w", e.Type, e.Version, err)
			}
			rec = Record{
				Name:      p.Name,
				UpdatedAt: p.UpdatedAt.UTC(),
				UpdatedBy: p.UpdatedBy,
				Version:   e.Version,
			}
		default:
			return Record{}, fmt.Errorf("unknown event %q", e.Type)
		}
	}
	return rec, nil
}
