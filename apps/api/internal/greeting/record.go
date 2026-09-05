package greeting

import "time"

// Record is the single greeting row shared by every client.
type Record struct {
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updatedAt"`
	UpdatedBy string    `json:"updatedBy"`
	Version   int64     `json:"version"`
}

// Store persists the singleton greeting.
type Store interface {
	Get() (Record, error)
	// SaveIfNewer writes incoming when it wins last-write-wins.
	// accepted is false when the current row is newer; current is always returned.
	SaveIfNewer(incoming Record) (current Record, accepted bool, err error)
	Ping() error
}
