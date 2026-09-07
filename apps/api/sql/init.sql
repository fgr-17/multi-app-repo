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

INSERT INTO users (email, given_name, family_name, display_name, phone, city, country, locale)
VALUES
  ('rouxfederico@gmail.com', 'Federico', 'Roux', 'Federico Roux', '+54 11 5555-0101', 'Buenos Aires', 'AR', 'es-AR'),
  ('ada@example.com', 'Ada', 'Lovelace', 'Ada Lovelace', '+44 20 7946 0001', 'Londres', 'GB', 'en-GB'),
  ('grace@example.com', 'Grace', 'Hopper', 'Grace Hopper', '+1 202 555 0147', 'Washington', 'US', 'en-US')
ON CONFLICT (email) DO NOTHING;
