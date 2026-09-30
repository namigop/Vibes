# Kanban

A Kanban board app: Go + PostgreSQL on the back, Next.js + TypeScript on the
front. Multiple boards, custom columns for deal stages, cards with priority /
assignee / due date, drag-and-drop between columns that saves on drop, and a
dark/light theme.

## Stack

| Layer     | Choice                                                    |
| --------- | --------------------------------------------------------- |
| Frontend  | Next.js 16 (App Router), React 19, TypeScript, Tailwind v4 |
| Drag-drop | `@dnd-kit/core` + `@dnd-kit/sortable`                     |
| Backend   | Go 1.27, `net/http` with the stdlib pattern router         |
| Database  | PostgreSQL 16 via `pgx/v5`, `sqlc`-generated queries      |

The SQL layer is generated from `backend/sql/queries.sql` by sqlc — no ORM, no
hand-written row scanning, and a compile error whenever a query and the schema
disagree.

## Layout

```
kanban/
├── backend/
│   ├── cmd/server/            # entrypoint: config -> pool -> migrate -> serve
│   ├── internal/
│   │   ├── api/               # HTTP handlers, DTOs, service layer, router, CORS
│   │   └── store/             # sqlc-generated queries + pool + demo seed
│   ├── migrations/            # embedded, applied at boot
│   ├── sql/                   # schema.sql + queries.sql (sqlc source of truth)
│   ├── test/smoke.ps1         # end-to-end API test against a running server
│   └── sqlc.yaml
├── frontend/                  # Next.js app
├── docs/API.md                # the frozen REST contract
└── docker-compose.yml
```

`docs/API.md` is the contract shared by both halves of the app. Change a shape
there and you change both sides.

## Running it

### Prerequisites

- Go 1.27+
- Node 22+ and npm 10+
- PostgreSQL 16 reachable
- Optional: sqlc, only if you want to regenerate the store
- Optional: Docker, only for the compose path

### 1. Database

Either point at an existing Postgres, or use the container:

```bash
cp .env.example .env
make db-up            # docker compose up -d postgres
```

Using a Postgres you already have instead? Just set `DATABASE_URL`.

### 2. Backend

```bash
cd backend
export DATABASE_URL="postgres://todo:todo@localhost:5432/kanban?sslmode=disable"
export BACKEND_PORT=8080
go run ./cmd/server
```

The server applies any pending migrations and seeds demo data on first boot —
but only if the database has no boards at all, so it will never touch real data.

Check it:

```bash
curl http://localhost:8080/api/v1/health
```

### 3. Frontend

```bash
cd frontend
cp .env.local.example .env.local   # or create .env.local
npm install
npm run dev
```

`NEXT_PUBLIC_API_BASE_URL` defaults to `http://localhost:8080/api/v1` and is
inlined at build time, so it must be set before `npm run build`.

Open http://localhost:3000.

### Everything at once

```bash
make up             # docker compose up --build
```

## Tests

```bash
make test           # Go unit tests (ordering, validation, date handling)
make smoke          # end-to-end: every endpoint against a running API
make check          # fmt, vet, test, frontend build, frontend lint
cd frontend && npm run parity   # frontend ordering logic vs. the live API
```

`make smoke` needs the API already running. It is safe to re-run: it cleans up
any board left behind by an interrupted run, and it asserts that column and
card positions stay a gap-free `0..n-1` run after every move.

`npm run parity` is the one that matters most for drag-and-drop. It feeds the
real `board-state.ts` helpers into a live API and checks that the index the UI
optimistically renders is the index the server actually settles on. "The card
lands one slot off after a cross-column drag" is invisible to tests on either
side in isolation.

> **Ports on this machine.** 8080, 5432 and 3000 are already taken by your
> `todo-app` stack and a native Postgres. The kanban backend was therefore run
> on **8081** and `frontend/.env.local` points at it. To go back to the
> defaults, stop the other stack and set `BACKEND_PORT=8080` plus
> `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1`, then rebuild the
> frontend (the value is inlined at build time).


## Design notes

**Ordering.** `position` is a gap-free ordinal within its parent (cards inside a
column, columns inside a board). A move reindexes both the source and the
destination inside one transaction, so a stale client can never corrupt the
order and there is no fractional-position drift to clean up. Out-of-range
positions are clamped, not rejected.

**Dates.** `cards.due_date` is a calendar date and is served as `YYYY-MM-DD`.
It is never round-tripped through a timestamp, so it cannot shift by a day for
anyone west of UTC. The frontend formats it locally for display only.

**The `date` type.** sqlc 1.31 does not honour a `date` override against the
pgx/v5 built-in mapping, so `due_date` stays `pgtype.Date` in the store. That
is deliberate: its `Valid` flag is exactly the nullable-calendar-date semantic
we want. See the note in `backend/sqlc.yaml`.

**Optimistic updates.** The frontend applies a drag immediately and reconciles
with the server afterwards, rolling back to the previous snapshot if the save
fails. See the frontend README for the details.

## API

Full reference in [`docs/API.md`](docs/API.md). In short:

```
GET    /api/v1/boards
POST   /api/v1/boards
GET    /api/v1/boards/{id}
PATCH  /api/v1/boards/{id}
DELETE /api/v1/boards/{id}

POST   /api/v1/boards/{id}/columns
PATCH  /api/v1/columns/{id}
POST   /api/v1/columns/{id}/move
DELETE /api/v1/columns/{id}

POST   /api/v1/columns/{id}/cards
PATCH  /api/v1/cards/{id}
POST   /api/v1/cards/{id}/move
DELETE /api/v1/cards/{id}

GET    /api/v1/health
```

## Configuration

| Variable               | Default                                          |
| ---------------------- | ------------------------------------------------ |
| `DATABASE_URL`         | `postgres://todo:todo@localhost:5432/kanban?sslmode=disable` |
| `BACKEND_PORT`         | `8080`                                           |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000`                          |
| `SEED_DEMO`            | `true` — seeds only when the database is empty   |
| `APP_ENV`              | `development`                                    |

In development every `localhost` / `127.0.0.1` port is additionally allowed
through CORS, so the Next dev server can change ports freely. Set `APP_ENV` to
anything else and only the configured origins are accepted.

## Regenerating the store

After editing `backend/sql/queries.sql` or `backend/sql/schema.sql`:

```bash
make install-sqlc    # once
make generate
```

`make generate` fails loudly if a query and the schema have drifted apart —
that is the whole point of using sqlc instead of writing SQL by hand.
