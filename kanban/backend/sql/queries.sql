-- name: ListBoards :many
SELECT *
FROM boards
ORDER BY position ASC, created_at ASC;

-- name: GetBoard :one
SELECT *
FROM boards
WHERE id = sqlc.arg(id)::uuid;

-- name: CreateBoard :one
INSERT INTO boards (name, position)
SELECT sqlc.arg(name)::text, COALESCE(MAX(position), -1) + 1
FROM boards
RETURNING *;

-- name: UpdateBoardName :one
UPDATE boards
SET name       = sqlc.arg(name)::text,
    updated_at = now()
WHERE id = sqlc.arg(id)::uuid
RETURNING *;

-- name: DeleteBoard :execrows
DELETE FROM boards
WHERE id = sqlc.arg(id)::uuid;

-- name: BoardExists :one
SELECT EXISTS (SELECT 1 FROM boards WHERE id = sqlc.arg(id)::uuid) AS exists;

-- name: ListColumns :many
SELECT *
FROM board_columns
WHERE board_id = sqlc.arg(board_id)::uuid
ORDER BY position ASC, created_at ASC;

-- name: ListColumnIDs :many
SELECT id
FROM board_columns
WHERE board_id = sqlc.arg(board_id)::uuid
ORDER BY position ASC, created_at ASC;

-- name: GetColumn :one
SELECT *
FROM board_columns
WHERE id = sqlc.arg(id)::uuid;

-- name: CreateColumn :one
INSERT INTO board_columns (board_id, name, position)
SELECT sqlc.arg(board_id)::uuid,
       sqlc.arg(name)::text,
       COALESCE(MAX(position), -1) + 1
FROM board_columns
WHERE board_id = sqlc.arg(board_id)::uuid
RETURNING *;

-- name: UpdateColumnName :one
UPDATE board_columns
SET name       = sqlc.arg(name)::text,
    updated_at = now()
WHERE id = sqlc.arg(id)::uuid
RETURNING *;

-- name: SetColumnOrdinal :exec
UPDATE board_columns
SET position   = sqlc.arg(position)::int,
    updated_at = now()
WHERE id = sqlc.arg(id)::uuid;

-- name: DeleteColumn :execrows
DELETE FROM board_columns
WHERE id = sqlc.arg(id)::uuid;

-- name: ListCardsForBoard :many
SELECT c.*
FROM cards c
         JOIN board_columns col ON col.id = c.column_id
WHERE col.board_id = sqlc.arg(board_id)::uuid
ORDER BY col.position ASC, c.position ASC, c.created_at ASC;

-- name: GetCard :one
SELECT *
FROM cards
WHERE id = sqlc.arg(id)::uuid;

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
UPDATE cards
SET title       = sqlc.arg(title)::text,
    description = COALESCE(sqlc.narg(description)::text, ''),
    assignee    = COALESCE(sqlc.narg(assignee)::text, ''),
    priority    = sqlc.arg(priority)::text,
    due_date    = sqlc.narg(due_date)::date,
    updated_at  = now()
WHERE id = sqlc.arg(id)::uuid
RETURNING *;

-- name: SetCardColumn :exec
UPDATE cards
SET column_id  = sqlc.arg(column_id)::uuid,
    updated_at = now()
WHERE id = sqlc.arg(id)::uuid;

-- name: SetCardPosition :exec
UPDATE cards
SET position   = sqlc.arg(position)::int,
    updated_at = now()
WHERE id = sqlc.arg(id)::uuid;

-- name: DeleteCard :execrows
DELETE FROM cards
WHERE id = sqlc.arg(id)::uuid;

-- name: ListCardIDsInColumn :many
SELECT id
FROM cards
WHERE column_id = sqlc.arg(column_id)::uuid
ORDER BY position ASC, created_at ASC;
