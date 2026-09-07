package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    given_name TEXT NOT NULL,
    family_name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    phone TEXT NOT NULL DEFAULT '',
    city TEXT NOT NULL DEFAULT '',
    country TEXT NOT NULL DEFAULT '',
    locale TEXT NOT NULL DEFAULT 'es-AR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const seed = `
INSERT INTO users (email, given_name, family_name, display_name, phone, city, country, locale)
VALUES
  ('rouxfederico@gmail.com', 'Federico', 'Roux', 'Federico Roux', '+54 11 5555-0101', 'Buenos Aires', 'AR', 'es-AR'),
  ('ada@example.com', 'Ada', 'Lovelace', 'Ada Lovelace', '+44 20 7946 0001', 'Londres', 'GB', 'en-GB'),
  ('grace@example.com', 'Grace', 'Hopper', 'Grace Hopper', '+1 202 555 0147', 'Washington', 'US', 'en-US')
ON CONFLICT (email) DO NOTHING;
`

type Postgres struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*Postgres, error) {
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
		return nil, fmt.Errorf("migrate users: %w", err)
	}
	if _, err := pool.Exec(ctx, seed); err != nil {
		pool.Close()
		return nil, fmt.Errorf("seed users: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() { p.pool.Close() }

func (p *Postgres) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return p.pool.Ping(ctx)
}

func (p *Postgres) List() ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := p.pool.Query(ctx, `
		SELECT id, email, given_name, family_name, display_name, phone, city, country, locale, created_at, updated_at
		FROM users
		ORDER BY family_name, given_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if out == nil {
		out = []User{}
	}
	return out, rows.Err()
}

func (p *Postgres) Get(id uuid.UUID) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	row := p.pool.QueryRow(ctx, `
		SELECT id, email, given_name, family_name, display_name, phone, city, country, locale, created_at, updated_at
		FROM users WHERE id = $1`, id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return u, err
}

func (p *Postgres) Create(in User) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if in.DisplayName == "" {
		in.DisplayName = in.GivenName + " " + in.FamilyName
	}
	if in.Locale == "" {
		in.Locale = "es-AR"
	}
	row := p.pool.QueryRow(ctx, `
		INSERT INTO users (email, given_name, family_name, display_name, phone, city, country, locale)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, email, given_name, family_name, display_name, phone, city, country, locale, created_at, updated_at`,
		in.Email, in.GivenName, in.FamilyName, in.DisplayName, in.Phone, in.City, in.Country, in.Locale)
	return scanUser(row)
}

var ErrNotFound = errors.New("user not found")

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (User, error) {
	var u User
	err := row.Scan(
		&u.ID, &u.Email, &u.GivenName, &u.FamilyName, &u.DisplayName,
		&u.Phone, &u.City, &u.Country, &u.Locale, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return User{}, err
	}
	u.CreatedAt = u.CreatedAt.UTC()
	u.UpdatedAt = u.UpdatedAt.UTC()
	return u, nil
}
