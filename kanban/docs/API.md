# Kanban API Contract (frozen)

This document is the **source of truth** shared by the Go backend and the
Next.js frontend. Do not change a shape here without updating both sides.

- Base URL (dev): `http://localhost:8080/api/v1`
- All request and response bodies are `application/json; charset=utf-8`.
- All IDs are UUIDv4 strings.
- Errors always use the envelope below, with an appropriate HTTP status.

## Authentication

Authentication is a **stateless bearer token**. There is no session table and no
cookie.

1. `POST /auth/login` with a username and password.
2. The server returns a signed **HS256 JWT** valid for `JWT_TTL` (24h default).
3. The client sends `Authorization: Bearer <token>` on **every** other request.
4. Logout is client-side only — the token stays valid until it expires. There is
   no revocation.

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

Claims are `sub` (the user id), `iss`, `aud`, `iat`, `nbf`, `exp`. The
username is deliberately **not** a claim: a rename must not invalidate
outstanding tokens, and a stale claim could otherwise show a name that no
longer exists.

### Accounts

There is **no signup endpoint**. Accounts are provisioned by the backend seed
(`backend/internal/store/seed.go`) as demo personas: `alice`, `bob`, `carol`,
`demo`, plus `legacy`. They all share `DEMO_PASSWORD` (default `kanban-demo`).

`legacy` is special: migration `00002_auth.sql` creates it and assigns it every
board that existed before ownership, so nothing predating the auth work is lost.

> Because every seeded account shares one published password, the server
> **refuses to start** when `APP_ENV=production` and `SEED_DEMO=true`, or when
> `JWT_SECRET` is still the development default. Anyone who can reach the URL
> could otherwise log in as any user.

### Ownership

A board belongs to **exactly one** user. There is no sharing.

Every board-scoped query filters on the authenticated user's id. Requesting
another user's board returns **404, not 403** — a 403 would confirm the board
exists, leaking the fact that someone has one.

### CSRF

Not applicable. The credential is an `Authorization` header, which the browser
never attaches automatically, so a cross-site request cannot borrow it. (This is
the trade for not using an `HttpOnly` cookie, which would have made the token
unreadable to JavaScript — and therefore impossible to send in a header.)

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
| 401    | `unauthorized`        | missing, malformed, expired or forged bearer token     |
| 404    | `not_found`           | board / column / card id does not exist **or belongs to someone else** |
| 409    | `conflict`            | duplicate name where uniqueness is required           |
| 500    | `internal_error`      | anything unexpected (details go to the server log)     |

A failed login returns `401` with the deliberately vague message
`invalid username or password` — identical for an unknown username and a wrong
password, so the endpoint cannot be used to enumerate accounts.

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
- **`assignee` is free text and is not a user reference.** Boards are owner-only,
  so the only person who can open a card is its owner; a user reference would
  name someone who by construction cannot see the card. It names a person
  involved in the work instead (a deal contact, a colleague with no account).

## Types

```jsonc
// User — never includes password_hash
{
  "id": "uuid",
  "username": "alice"
}

// LoginResponse
{
  "token": "eyJhbGciOi...",
  "token_type": "Bearer",
  "expires_at": "2026-10-14T05:35:44Z",
  "user": { /* User */ }
}

// Board
{
  "id": "uuid",
  "name": "string",
  "position": 0,
  "owner_id": "uuid",           // the only user who can see this board
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
  "assignee": "string",       // free text, NOT a user — see Domain rules
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

Unless marked **public**, every endpoint requires
`Authorization: Bearer <token>`.

### Auth

| Method | Path           | Auth   | Body                            | Response         |
| ------ | -------------- | ------ | ------------------------------- | ---------------- |
| POST   | `/auth/login`  | public | `{ "username", "password" }`    | `LoginResponse`  |
| GET    | `/auth/me`     | bearer | –                               | `User`           |

### Boards

| Method | Path                | Body                     | Response          |
| ------ | ------------------- | ------------------------ | ----------------- |
| GET    | `/boards`           | –                        | `Board[]`         |
| POST   | `/boards`           | `{ "name": "string" }`   | `Board` 201       |
| GET    | `/boards/{id}`      | –                        | `BoardDetail`     |
| PATCH  | `/boards/{id}`      | `{ "name": "string" }`   | `Board`           |
| DELETE | `/boards/{id}`      | –                        | `204` empty       |

`GET /boards` returns only the caller's boards. `DELETE /boards/{id}` cascades
to that board's columns and cards.

### Columns

| Method | Path                            | Body                     | Response      |
| ------ | ------------------------------- | ------------------------ | ------------- |
| POST   | `/boards/{boardId}/columns`     | `{ "name": "string" }`   | `Column` 201  |
| PATCH  | `/columns/{id}`                 | `{ "name": "string" }`   | `Column`      |
| POST   | `/columns/{id}/move`            | `{ "position": 0..n-1 }` | `Column[]`    |
| DELETE | `/columns/{id}`                 | –                        | `204` empty   |

`POST /columns/{id}/move` reindexes **all** columns of that board and returns
them in their new order, so the client can replace local state wholesale.

`DELETE /columns/{id}` cascades to the cards inside it. A board may be left
with zero columns; the UI must render that state gracefully.

### Cards

| Method | Path                  | Body                                          | Response    |
| ------ | --------------------- | --------------------------------------------- | ----------- |
| POST   | `/columns/{id}/cards` | `CreateCardRequest`                           | `Card` 201  |
| PATCH  | `/cards/{id}`         | `CreateCardRequest` (full replace, all fields)  | `Card`      |
| POST   | `/cards/{id}/move`    | `{ "column_id": "uuid", "position": 0..n-1 }`  | `Card`      |
| DELETE | `/cards/{id}`         | –                                             | `204` empty |

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

| Method | Path     | Auth   | Response                        |
| ------ | -------- | ------ | ------------------------------- |
| GET    | `/health` | public | `{ "status": "ok", "db": "ok" }` |

## CORS

Development only. The server allows the origin in `CORS_ALLOWED_ORIGINS`
(default `http://localhost:3000`), plus any `localhost`/`127.0.0.1` port, and
exposes the `Content-Type` header. Credentials are not used.

## Environment variables

| Variable               | Default                                                            |
| ---------------------- | ------------------------------------------------------------------ |
| `DATABASE_URL`         | `postgres://kanban:kanban@postgres:5432/kanban?sslmode=disable` (compose network) |
| `BACKEND_PORT`         | `8080`                                                             |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000`                                            |
| `SEED_DEMO`            | `true` — seed demo accounts only when the database has no users     |
| `APP_ENV`              | `development`                                                      |
| `JWT_SECRET`           | a built-in development default; **required** and ≥32 chars in production |
| `JWT_TTL`              | `24h`                                                              |
| `DEMO_PASSWORD`        | `kanban-demo`                                                      |

Run `go run ./cmd/hashgen <password>` to produce a bcrypt digest — needed if
you ever hardcode one in a migration.
