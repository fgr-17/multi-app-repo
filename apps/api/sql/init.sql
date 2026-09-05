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
