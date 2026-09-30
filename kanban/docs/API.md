# Kanban API Contract (frozen)

This document is the **source of truth** shared by the Go backend and the
Next.js frontend. Do not change a shape here without updating both sides.

- Base URL (dev): `http://localhost:8080/api/v1`
- All request and response bodies are `application/json; charset=utf-8`.
- All IDs are UUIDv4 strings.
- Errors always use the envelope below, with an appropriate HTTP status.

## Error envelope

```jsonc
{
  "error": "validation_failed",        // stable machine-readable slug
  "message": "title must not be empty" // human-readable, safe to show in a toast
}
```

| Status | `error` slug          | When                                                  |
| ------ | --------------------- | ----------------------------------------------------- |
| 400    | `validation_failed`   | malformed body, bad UUID, unknown priority value      |
| 404    | `not_found`           | board / column / card id does not exist               |
| 409    | `conflict`            | duplicate name where uniqueness is required           |
| 500    | `internal_error`      | anything unexpected (details go to the server log)     |

## Domain rules

- `priority` ∈ `low` | `medium` | `high` | `urgent`. Default `medium`.
- `position` is a **gap-free 0..n-1 ordinal** scoped to its parent
  (cards within a column, columns within a board). The server reindexes the
  source and destination on every move, so clients never send fractional
  positions and the two can never drift.
- `due_date` is a **calendar date** rendered as `"YYYY-MM-DD"`, or `null`.
  It is never a timestamp and must never be shifted by the client's timezone.
- Empty string and `null` are equivalent for `description` / `assignee`;
  the server normalises to `""`.
- Titles and names must be non-empty after trimming.

## Types

```jsonc
// Board
{
  "id": "uuid",
  "name": "string",
  "position": 0,
  "created_at": "2026-09-29T08:00:00Z", // RFC3339 UTC
  "updated_at": "2026-09-29T08:00:00Z"
}

// Column
{
  "id": "uuid",
  "board_id": "uuid",
  "name": "string",
  "position": 0,
  "created_at": "2026-09-29T08:00:00Z",
  "updated_at": "2026-09-29T08:00:00Z"
}

// Card
{
  "id": "uuid",
  "column_id": "uuid",
  "title": "string",
  "description": "string",   // "" when unset
  "assignee": "string",       // "" when unassigned
  "priority": "low|medium|high|urgent",
  "due_date": "2026-10-15",  // or null
  "position": 0,
  "created_at": "2026-09-29T08:00:00Z",
  "updated_at": "2026-09-29T08:00:00Z"
}

// BoardDetail — everything needed to render one board in a single request
{
  "board": { /* Board */ },
  "columns": [ { /* Column */ } ],          // ordered by position
  "cards":   [ { /* Card */ } ]             // ordered by (column.position, position)
}
```

## Endpoints

### Boards

| Method | Path                | Body                          | Response          |
| ------ | ------------------- | ----------------------------- | ----------------- |
| GET    | `/boards`           | –                             | `Board[]`         |
| POST   | `/boards`           | `{ "name": "string" }`        | `Board` 201       |
| GET    | `/boards/{id}`      | –                             | `BoardDetail`     |
| PATCH  | `/boards/{id}`      | `{ "name": "string" }`        | `Board`           |
| DELETE | `/boards/{id}`      | –                             | `204` empty       |

`DELETE /boards/{id}` cascades to that board's columns and cards.

### Columns

| Method | Path                            | Body                            | Response      |
| ------ | ------------------------------- | ------------------------------- | ------------- |
| POST   | `/boards/{boardId}/columns`     | `{ "name": "string" }`          | `Column` 201  |
| PATCH  | `/columns/{id}`                 | `{ "name": "string" }`          | `Column`      |
| POST   | `/columns/{id}/move`            | `{ "position": 0..n-1 }`        | `Column[]`    |
| DELETE | `/columns/{id}`                 | –                               | `204` empty   |

`POST /columns/{id}/move` reindexes **all** columns of that board and returns
them in their new order, so the client can replace local state wholesale.

`DELETE /columns/{id}` cascades to the cards inside it. A board may be left
with zero columns; the UI must render that state gracefully.

### Cards

| Method | Path                  | Body                                                  | Response    |
| ------ | --------------------- | ----------------------------------------------------- | ----------- |
| POST   | `/columns/{id}/cards` | `CreateCardRequest`                                   | `Card` 201  |
| PATCH  | `/cards/{id}`         | `CreateCardRequest` (full replace, all fields)        | `Card`      |
| POST   | `/cards/{id}/move`    | `{ "column_id": "uuid", "position": 0..n-1 }`         | `Card`      |
| DELETE | `/cards/{id}`         | –                                                     | `204` empty |

```jsonc
// CreateCardRequest — used verbatim for both create and full update
{
  "title": "string",         // required, non-empty after trim
  "description": "string",   // optional, default ""
  "assignee": "string",      // optional, default ""
  "priority": "medium",      // optional, default "medium"
  "due_date": "2026-10-15"   // optional, nullable
}
```

`POST /cards/{id}/move` semantics:

1. `position` is the desired **index within the destination column after the move**.
2. Values outside `0..len` are clamped rather than rejected.
3. If `column_id` equals the card's current column, it is a reorder in place.
4. The server reindexes both the source and destination column in one
   transaction and returns the moved card with its settled `position` and
   `column_id`.

### Health

| Method | Path     | Response                        |
| ------ | -------- | ------------------------------- |
| GET    | `/health` | `{ "status": "ok", "db": "ok" }` |

## CORS

Development only. The server allows the origin in `CORS_ALLOWED_ORIGINS`
(default `http://localhost:3000`), plus any `localhost`/`127.0.0.1` port, and
exposes the `Content-Type` header. Credentials are not used.

## Environment variables

| Variable              | Default                                        |
| --------------------- | ---------------------------------------------- |
| `DATABASE_URL`        | `postgres://todo:todo@localhost:5432/kanban?sslmode=disable` |
| `BACKEND_PORT`        | `8080`                                         |
| `CORS_ALLOWED_ORIGINS`| `http://localhost:3000`                        |
| `SEED_DEMO`           | `true` — seed demo data only when the DB has no boards |
| `APP_ENV`             | `development`                                  |
