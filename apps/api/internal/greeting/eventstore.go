package greeting

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const eventSchema = `
CREATE TABLE IF NOT EXISTS events (
    stream_id TEXT NOT NULL,
    version BIGINT NOT NULL,
    type TEXT NOT NULL,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (stream_id, version)
);

CREATE TABLE IF NOT EXISTS outbox (
    id BIGSERIAL PRIMARY KEY,
    stream_id TEXT NOT NULL,
    version BIGINT NOT NULL,
    payload JSONB NOT NULL,
    published BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS outbox_unpublished_idx ON outbox (id) WHERE NOT published;
`

// EventStore is the write model: append-only events in Postgres + outbox for Kafka.
type EventStore struct {
	pool *pgxpool.Pool
}

func OpenEventStore(ctx context.Context, databaseURL string) (*EventStore, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if _, err := pool.Exec(ctx, eventSchema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	es := &EventStore{pool: pool}
	if err := es.seed(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return es, nil
}

func (s *EventStore) Close() { s.pool.Close() }

func (s *EventStore) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.pool.Ping(ctx)
}

func (s *EventStore) List() ([]Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.load(ctx, nil)
}

func (s *EventStore) Get() (Record, error) {
	events, err := s.List()
	if err != nil {
		return Record{}, err
	}
	return Fold(events)
}

func (s *EventStore) SaveIfNewer(incoming Record) (Record, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Record{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	events, err := s.load(ctx, tx)
	if err != nil {
		return Record{}, false, err
	}
	current, err := Fold(events)
	if err != nil {
		return Record{}, false, err
	}
	if !Wins(incoming, current) {
		if err := tx.Commit(ctx); err != nil {
			return Record{}, false, err
		}
		return current, false, nil
	}

	next := incoming
	next.Version = current.Version + 1
	if next.UpdatedAt.IsZero() {
		next.UpdatedAt = time.Now().UTC()
	}
	evt, err := NewRenamed(next.Version, next, TypeGreetingRenamed)
	if err != nil {
		return Record{}, false, err
	}
	if err := appendEvent(ctx, tx, evt); err != nil {
		return Record{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Record{}, false, err
	}
	return next, true, nil
}

func (s *EventStore) Unpublished(ctx context.Context, limit int) ([]Event, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT stream_id, version, payload
		FROM outbox
		WHERE NOT published
		ORDER BY id
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.StreamID, &e.Version, &e.Payload); err != nil {
			return nil, err
		}
		var envelope Event
		if err := json.Unmarshal(e.Payload, &envelope); err != nil {
			return nil, err
		}
		out = append(out, envelope)
	}
	return out, rows.Err()
}

func (s *EventStore) MarkPublished(ctx context.Context, streamID string, version int64) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE outbox SET published = TRUE
		WHERE stream_id = $1 AND version = $2 AND NOT published`, streamID, version)
	return err
}

func (s *EventStore) seed(ctx context.Context) error {
	var n int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM events WHERE stream_id = $1`, StreamID).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	rec := Record{Name: "Mundo", UpdatedAt: time.Now().UTC(), UpdatedBy: "seed", Version: 1}
	evt, err := NewRenamed(1, rec, TypeGreetingSeeded)
	if err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := appendEvent(ctx, tx, evt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *EventStore) load(ctx context.Context, tx pgx.Tx) ([]Event, error) {
	q := `
		SELECT stream_id, version, type, payload, occurred_at
		FROM events
		WHERE stream_id = $1
		ORDER BY version`
	var rows pgx.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(ctx, q, StreamID)
	} else {
		rows, err = s.pool.Query(ctx, q, StreamID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.StreamID, &e.Version, &e.Type, &e.Payload, &e.OccurredAt); err != nil {
			return nil, err
		}
		e.OccurredAt = e.OccurredAt.UTC()
		events = append(events, e)
	}
	return events, rows.Err()
}

func appendEvent(ctx context.Context, tx pgx.Tx, evt Event) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO events (stream_id, version, type, payload, occurred_at)
		VALUES ($1, $2, $3, $4, $5)`,
		evt.StreamID, evt.Version, evt.Type, evt.Payload, evt.OccurredAt)
	if err != nil {
		return err
	}
	envelope, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox (stream_id, version, payload)
		VALUES ($1, $2, $3)`, evt.StreamID, evt.Version, envelope)
	return err
}

func DatabaseURL() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://hola:hola@localhost:5432/hola?sslmode=disable"
}
