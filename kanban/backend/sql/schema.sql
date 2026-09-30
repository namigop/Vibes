-- Kanban board schema.
-- Source of truth for sqlc code generation (see backend/sqlc.yaml).
-- The identical DDL is applied at runtime from backend/migrations/00001_init.sql.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- A board is one pipeline of columns (e.g. a deal stage tracker).
CREATE TABLE boards (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL CHECK (length(btrim(name)) > 0),
    position   INTEGER     NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Named "board_columns" rather than "columns": the latter collides with
-- reserved SQL keywords in some dialects and forces quoting everywhere.
-- One column == one workflow stage (To Do, In Progress, Review, Done, ...).
CREATE TABLE board_columns (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id   UUID        NOT NULL REFERENCES boards (id) ON DELETE CASCADE,
    name       TEXT        NOT NULL CHECK (length(btrim(name)) > 0),
    position   INTEGER     NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_board_columns_board_position
    ON board_columns (board_id, position);

-- A card is one unit of work sitting in exactly one column.
-- `position` is a gap-free 0..n-1 ordinal inside its column. Moves reindex
-- both the source and the destination column in a single transaction, so the
-- ordinals can never drift out of sync.
CREATE TABLE cards (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    column_id   UUID        NOT NULL REFERENCES board_columns (id) ON DELETE CASCADE,
    title       TEXT        NOT NULL CHECK (length(btrim(title)) > 0),
    description TEXT        NOT NULL DEFAULT '',
    assignee    TEXT        NOT NULL DEFAULT '',
    priority    TEXT        NOT NULL DEFAULT 'medium'
                 CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    due_date    DATE,
    position    INTEGER     NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_cards_column_position
    ON cards (column_id, position);
