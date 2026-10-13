-- Authentication and per-user board ownership.
--
-- Applied by the built-in migration runner
-- (backend/internal/platform/database/migrations/migrate.go), which records this
-- filename in schema_migrations. Each file runs inside one transaction, so a
-- failure here leaves the database on the last good version.

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      TEXT        NOT NULL CHECK (length(btrim(username)) > 0),
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Case-insensitive uniqueness: "Alice" and "alice" are the same account.
-- A plain UNIQUE(username) would happily store both.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_lower
    ON users (lower(username));

-- The `legacy` account adopts boards that existed before ownership existed.
--
-- This hash is bcrypt("kanban-demo"), the default demo password. It is a
-- literal because bcrypt salts are random and cannot be computed in SQL; the
-- Go seed rewrites it from DEMO_PASSWORD on boot when seeding is enabled, so
-- overriding the password still works.
--
-- Regenerate with:  go run ./cmd/hashgen "kanban-demo"
INSERT INTO users (username, password_hash)
VALUES ('legacy', '$2a$10$ON1u6glM.chKsz7PnDUQ2O.UCPJ2EvE9oTT3JsLbDDURWuSkJjwFK')
ON CONFLICT DO NOTHING;

-- Added as nullable first so existing rows can be backfilled before the
-- constraint is enforced.
ALTER TABLE boards ADD COLUMN IF NOT EXISTS owner_id UUID REFERENCES users (id) ON DELETE CASCADE;

-- Every pre-auth board belongs to `legacy`. Log in as legacy to see them.
UPDATE boards
SET owner_id = (SELECT id FROM users WHERE lower(username) = 'legacy')
WHERE owner_id IS NULL;

-- A board with no owner cannot be reached by any authenticated user, so it
-- would be invisible data. Fail the migration instead of creating that state.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM boards WHERE owner_id IS NULL) THEN
        RAISE EXCEPTION 'cannot enforce NOT NULL: % board(s) still have no owner',
            (SELECT count(*) FROM boards WHERE owner_id IS NULL);
    END IF;
END
$$;

ALTER TABLE boards ALTER COLUMN owner_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_boards_owner_position
    ON boards (owner_id, position);