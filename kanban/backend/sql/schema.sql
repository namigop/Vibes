-- Kanban board schema.
-- Source of truth for sqlc code generation (see backend/sqlc.yaml).
-- The identical DDL is applied at runtime from
-- backend/internal/platform/database/migrations/00001_init.sql.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- An account. There is no signup route: accounts are provisioned by the seed
-- (see backend/internal/devseed/seed.go) so the app ships with a fixed set of
-- demo personas.
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      TEXT        NOT NULL CHECK (length(btrim(username)) > 0),
    -- bcrypt digest. The plaintext is never stored.
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Usernames are compared case-insensitively ("Alice" must not be a different
-- account from "alice"), so uniqueness is enforced on lower(username).
CREATE UNIQUE INDEX idx_users_username_lower ON users (lower(username));

-- A board is one pipeline of columns (e.g. a deal stage tracker).
CREATE TABLE boards (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL CHECK (length(btrim(name)) > 0),
    position   INTEGER     NOT NULL DEFAULT 0,
    -- Every board belongs to exactly one user. NOT NULL plus the
    -- ON DELETE CASCADE means there is no such thing as an unowned board, and
    -- deleting a user takes their boards with them.
    owner_id   UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_boards_owner_position
    ON boards (owner_id, position);

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
    -- Free text, deliberately NOT a foreign key to users.
    --
    -- Boards are owner-only, so the only person who can ever open a given card
    -- is its owner. A user reference here would therefore name someone who by
    -- construction cannot see the card. Instead, assignee names a person
    -- involved in the work being tracked (a deal contact, a colleague who has
    -- no account) and nothing more.
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
