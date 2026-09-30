-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE todos (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    done        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX todos_created_at_idx ON todos (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS todos;
