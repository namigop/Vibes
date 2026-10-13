-- Queries for the kanban app.
--
-- OWNERSHIP RULE: every query that can reach a board, column or card is scoped
-- to the authenticated user. A board is visible only to its owner, so
-- columns and cards are scoped by joining up to boards.owner_id rather than
-- trusting an id from the client.
--
-- A miss here is what makes another user's board return 404 rather than 403:
-- the row simply is not visible, so pgx reports ErrNoRows and the API maps it
-- to not_found. That is deliberate — a 403 would confirm the board exists.

-- ---------------------------------------------------------------------------
-- users / auth
-- ---------------------------------------------------------------------------

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = sqlc.arg(id)::uuid;

-- name: GetUserByUsername :one
-- Case-insensitive: login is case-insensitive even though display is not.
SELECT *
FROM users
WHERE lower(username) = lower(sqlc.arg(username)::text);

-- name: CreateUser :one
INSERT INTO users (username, password_hash)
VALUES (sqlc.arg(username)::text, sqlc.arg(password_hash)::text)
RETURNING *;

-- name: CountUsers :one
SELECT count(*)::int AS count
FROM users;

-- name: UpsertPersona :one
-- Used only by the seed. ON CONFLICT is not available on an expression index,
-- so it matches the existing row by lower(username) explicitly.
INSERT INTO users (username, password_hash)
SELECT sqlc.arg(username)::text, sqlc.arg(password_hash)::text
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE lower(username) = lower(sqlc.arg(username)::text)
)
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = sqlc.arg(password_hash)::text,
    updated_at    = now()
WHERE id = sqlc.arg(id)::uuid;

-- ---------------------------------------------------------------------------
-- boards
-- ---------------------------------------------------------------------------

-- name: ListBoards :many
SELECT *
FROM boards
WHERE owner_id = sqlc.arg(owner_id)::uuid
ORDER BY position ASC, created_at ASC;

-- name: GetBoard :one
SELECT *
FROM boards
WHERE id = sqlc.arg(id)::uuid
  AND owner_id = sqlc.arg(owner_id)::uuid;

-- name: CreateBoard :one
INSERT INTO boards (name, position, owner_id)
SELECT sqlc.arg(name)::text,
       COALESCE(MAX(position), -1) + 1,
       sqlc.arg(owner_id)::uuid
FROM boards
WHERE owner_id = sqlc.arg(owner_id)::uuid
RETURNING *;

-- name: UpdateBoardName :one
UPDATE boards
SET name       = sqlc.arg(name)::text,
    updated_at = now()
WHERE id = sqlc.arg(id)::uuid
  AND owner_id = sqlc.arg(owner_id)::uuid
RETURNING *;

-- name: DeleteBoard :execrows
DELETE FROM boards
WHERE id = sqlc.arg(id)::uuid
  AND owner_id = sqlc.arg(owner_id)::uuid;

-- name: BoardExists :one
SELECT EXISTS (
    SELECT 1
    FROM boards
    WHERE id = sqlc.arg(id)::uuid
      AND owner_id = sqlc.arg(owner_id)::uuid
) AS exists;

-- ---------------------------------------------------------------------------
-- columns
-- ---------------------------------------------------------------------------

-- name: ListColumns :many
SELECT col.*
FROM board_columns col
         JOIN boards b ON b.id = col.board_id
WHERE col.board_id = sqlc.arg(board_id)::uuid
  AND b.owner_id = sqlc.arg(owner_id)::uuid
ORDER BY col.position ASC, col.created_at ASC;

-- name: ListColumnIDs :many
SELECT col.id
FROM board_columns col
         JOIN boards b ON b.id = col.board_id
WHERE col.board_id = sqlc.arg(board_id)::uuid
  AND b.owner_id = sqlc.arg(owner_id)::uuid
ORDER BY col.position ASC, col.created_at ASC;

-- name: GetColumn :one
SELECT col.*
FROM board_columns col
         JOIN boards b ON b.id = col.board_id
WHERE col.id = sqlc.arg(id)::uuid
  AND b.owner_id = sqlc.arg(owner_id)::uuid;

-- name: CreateColumn :one
INSERT INTO board_columns (board_id, name, position)
SELECT sqlc.arg(board_id)::uuid,
       sqlc.arg(name)::text,
       COALESCE(MAX(position), -1) + 1
FROM board_columns
WHERE board_id = sqlc.arg(board_id)::uuid
RETURNING *;

-- name: UpdateColumnName :one
UPDATE board_columns col
SET name       = sqlc.arg(name)::text,
    updated_at = now()
FROM boards b
WHERE col.id = sqlc.arg(id)::uuid
  AND b.id = col.board_id
  AND b.owner_id = sqlc.arg(owner_id)::uuid
RETURNING col.*;

-- name: SetColumnOrdinal :exec
UPDATE board_columns col
SET position   = sqlc.arg(position)::int,
    updated_at = now()
FROM boards b
WHERE col.id = sqlc.arg(id)::uuid
  AND b.id = col.board_id
  AND b.owner_id = sqlc.arg(owner_id)::uuid;

-- name: DeleteColumn :execrows
DELETE FROM board_columns col
    USING boards b
WHERE col.id = sqlc.arg(id)::uuid
  AND b.id = col.board_id
  AND b.owner_id = sqlc.arg(owner_id)::uuid;

-- ---------------------------------------------------------------------------
-- cards
-- ---------------------------------------------------------------------------

-- name: ListCardsForBoard :many
SELECT c.*
FROM cards c
         JOIN board_columns col ON col.id = c.column_id
         JOIN boards b ON b.id = col.board_id
WHERE b.id = sqlc.arg(board_id)::uuid
  AND b.owner_id = sqlc.arg(owner_id)::uuid
ORDER BY col.position ASC, c.position ASC, c.created_at ASC;

-- name: GetCard :one
SELECT c.*
FROM cards c
         JOIN board_columns col ON col.id = c.column_id
         JOIN boards b ON b.id = col.board_id
WHERE c.id = sqlc.arg(id)::uuid
  AND b.owner_id = sqlc.arg(owner_id)::uuid;

-- name: CreateCard :one
INSERT INTO cards (column_id, title, description, assignee, priority, due_date, position)
SELECT sqlc.arg(column_id)::uuid,
       sqlc.arg(title)::text,
       COALESCE(sqlc.narg(description)::text, ''),
       COALESCE(sqlc.narg(assignee)::text, ''),
       sqlc.arg(priority)::text,
       sqlc.narg(due_date)::date,
       COALESCE(MAX(position), -1) + 1
FROM cards
WHERE column_id = sqlc.arg(column_id)::uuid
RETURNING *;

-- name: UpdateCard :one
UPDATE cards c
SET title       = sqlc.arg(title)::text,
    description = COALESCE(sqlc.narg(description)::text, ''),
    assignee    = COALESCE(sqlc.narg(assignee)::text, ''),
    priority    = sqlc.arg(priority)::text,
    due_date    = sqlc.narg(due_date)::date,
    updated_at  = now()
FROM board_columns col, boards b
WHERE c.id = sqlc.arg(id)::uuid
  AND col.id = c.column_id
  AND b.id = col.board_id
  AND b.owner_id = sqlc.arg(owner_id)::uuid
RETURNING c.*;

-- name: SetCardColumn :exec
UPDATE cards c
SET column_id  = sqlc.arg(column_id)::uuid,
    updated_at = now()
FROM board_columns col, boards b
WHERE c.id = sqlc.arg(id)::uuid
  AND col.id = c.column_id
  AND b.id = col.board_id
  AND b.owner_id = sqlc.arg(owner_id)::uuid;

-- name: SetCardPosition :exec
UPDATE cards c
SET position   = sqlc.arg(position)::int,
    updated_at = now()
FROM board_columns col, boards b
WHERE c.id = sqlc.arg(id)::uuid
  AND col.id = c.column_id
  AND b.id = col.board_id
  AND b.owner_id = sqlc.arg(owner_id)::uuid;

-- name: DeleteCard :execrows
DELETE FROM cards c
    USING board_columns col, boards b
WHERE c.id = sqlc.arg(id)::uuid
  AND col.id = c.column_id
  AND b.id = col.board_id
  AND b.owner_id = sqlc.arg(owner_id)::uuid;

-- name: ListCardIDsInColumn :many
SELECT c.id
FROM cards c
         JOIN board_columns col ON col.id = c.column_id
         JOIN boards b ON b.id = col.board_id
WHERE c.column_id = sqlc.arg(column_id)::uuid
  AND b.owner_id = sqlc.arg(owner_id)::uuid
ORDER BY c.position ASC, c.created_at ASC;