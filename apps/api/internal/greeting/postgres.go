package greeting

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schema = `
CREATE TABLE IF NOT EXISTS greeting (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    name TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    updated_by TEXT NOT NULL,
    version BIGINT NOT NULL
);
INSERT INTO greeting (id, name, updated_at, updated_by, version)
VALUES (1, 'Mundo', NOW(), 'seed', 1)
ON CONFLICT (id) DO NOTHING;
`

// Postgres is the production store. The greeting is a singleton row (id = 1).
type Postgres struct {
	pool *pgxpool.Pool
}

func OpenPostgres(ctx context.Context, databaseURL string) (*Postgres, error) {
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
	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() { p.pool.Close() }

func (p *Postgres) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return p.pool.Ping(ctx)
}

func (p *Postgres) Get() (Record, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanRow(p.pool.QueryRow(ctx, `
		SELECT name, updated_at, updated_by, version
		FROM greeting WHERE id = 1`))
}

func (p *Postgres) SaveIfNewer(incoming Record) (Record, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Record{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	current, err := scanRow(tx.QueryRow(ctx, `
		SELECT name, updated_at, updated_by, version
		FROM greeting WHERE id = 1
		FOR UPDATE`))
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

	row := tx.QueryRow(ctx, `
		UPDATE greeting
		SET name = $1, updated_at = $2, updated_by = $3, version = $4
		WHERE id = 1
		RETURNING name, updated_at, updated_by, version`,
		next.Name, next.UpdatedAt, next.UpdatedBy, next.Version)
	saved, err := scanRow(row)
	if err != nil {
		return Record{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Record{}, false, err
	}
	return saved, true, nil
}

func scanRow(row interface{ Scan(dest ...any) error }) (Record, error) {
	var rec Record
	if err := row.Scan(&rec.Name, &rec.UpdatedAt, &rec.UpdatedBy, &rec.Version); err != nil {
		return Record{}, err
	}
	rec.UpdatedAt = rec.UpdatedAt.UTC()
	return rec, nil
}

func DatabaseURL() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://hola:hola@localhost:5432/hola?sslmode=disable"
}
