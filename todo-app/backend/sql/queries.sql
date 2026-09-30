-- name: ListTodos :many
SELECT id, title, description, done, created_at, updated_at
FROM todos
ORDER BY created_at DESC;

-- name: GetTodo :one
SELECT id, title, description, done, created_at, updated_at
FROM todos
WHERE id = $1;

-- name: CreateTodo :one
INSERT INTO todos (title, description)
VALUES ($1, $2)
RETURNING id, title, description, done, created_at, updated_at;

-- name: UpdateTodo :one
UPDATE todos
SET
    title       = COALESCE(sqlc.narg('title'),       title),
    description = COALESCE(sqlc.narg('description'), description),
    done        = COALESCE(sqlc.narg('done'),        done),
    updated_at  = NOW()
WHERE id = sqlc.arg('id')
RETURNING id, title, description, done, created_at, updated_at;

-- name: DeleteTodo :execrows
DELETE FROM todos
WHERE id = $1;
