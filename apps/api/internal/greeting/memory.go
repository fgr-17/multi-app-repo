package greeting

import (
	"fmt"
	"sync"
	"time"
)

// Memory is an in-process store used by unit tests.
type Memory struct {
	mu    sync.Mutex
	row   Record
	ready bool
}

func NewMemory(seed Record) *Memory {
	return &Memory{row: seed, ready: true}
}

func (m *Memory) Get() (Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.ready {
		return Record{}, fmt.Errorf("empty store")
	}
	return m.row, nil
}

func (m *Memory) SaveIfNewer(incoming Record) (Record, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.ready || Wins(incoming, m.row) {
		incoming.Version = m.row.Version + 1
		if incoming.UpdatedAt.IsZero() {
			incoming.UpdatedAt = time.Now().UTC()
		}
		m.row = incoming
		m.ready = true
		return m.row, true, nil
	}
	return m.row, false, nil
}

func (m *Memory) Ping() error { return nil }
