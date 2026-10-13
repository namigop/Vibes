# Kanban

A Kanban board app: Go + PostgreSQL on the back, Next.js + TypeScript on the
front. Multiple boards, custom columns for deal stages, cards with priority /
assignee / due date, drag-and-drop between columns that saves on drop, and a
dark/light theme. Boards are private to their owner and access requires signing
in.

## Stack

| Layer     | Choice                                                    |
| --------- | --------------------------------------------------------- |
| Frontend  | Next.js 16 (App Router), React 19, TypeScript, Tailwind v4 |
| Drag-drop | `@dnd-kit/core` + `@dnd-kit/sortable`                     |
| Backend   | Go 1.27, `net/http` with the stdlib pattern router         |
| Auth      | bcrypt + stateless HS256 JWT bearer tokens                |
| Database  | PostgreSQL 16 via `pgx/v5`, `sqlc`-generated queries      |

The SQL layer is generated from `backend/sql/queries.sql` by sqlc — no ORM, no
hand-written row scanning, and a compile error whenever a query and the schema
disagree.

## Accounts

There is **no sign-up**. The backend seeds demo accounts on first boot:

| Username | Boards                                            |
| -------- | ------------------------------------------------- |
| `alice`  | Product Roadmap, Product Roadmap (empty)          |
| `bob`    | Sales Pipeline, Sales Pipeline (empty)            |
| `carol`  | Marketing Launch, Marketing Launch (empty)        |
| `demo`   | Getting Started, Getting Started (empty)          |
| `legacy` | any board that existed before ownership was added |

All of them share the password **`kanban-demo`** (override with
`DEMO_PASSWORD`). Sign in as each one to see that boards really are private to
their owner.

> **This is a development affordance, not a security model.** There is no
> signup flow, so every account comes from the seed and every account shares one
> published password — anyone who can reach the URL can log in as anyone. The
> server enforces this structurally: it **refuses to start** when
> `APP_ENV=production` while `SEED_DEMO=true`, while `JWT_SECRET` is still the
> built-in default, or while `DEMO_PASSWORD` is still the default.

See [docs/API.md](docs/API.md) for the authentication contract.

## Layout

```
kanban/
├── backend/
│   ├── cmd/server/            # composition root: config -> pool -> migrate -> register slices -> serve
│   ├── internal/
│   │   ├── platform/          # infrastructure, no business rules
│   │   │   ├── config/        # environment variables with defaults + production guard
│   │   │   ├── auth/          # bcrypt, JWT HS256, bearer middleware
│   │   │   ├── database/      # pool + UnitOfWork; migrations/ embedded and applied at boot
│   │   │   └── httpserver/    # router (Public/Private), CORS, JSON error envelope, decode
│   │   ├── domain/            # entities, errors, validation and ordering invariants
│   │   ├── features/          # vertical slices, one folder per resource
│   │   │   ├── auth/          #   login, me
│   │   │   ├── boards/        #   list, get, create, rename, delete
│   │   │   ├── columns/       #   create, rename, move, delete
│   │   │   └── cards/         #   create, update, move, delete
│   │   ├── store/             # sqlc-generated (horizontal on purpose — see below)
│   │   └── devseed/           # demo personas and sample boards
│   ├── sql/                   # schema.sql + queries.sql (sqlc source of truth)
│   ├── test/smoke.ps1         # end-to-end API test against a running server
│   └── sqlc.yaml
├── frontend/                  # Next.js app
├── docs/API.md                # the frozen REST contract
└── docker-compose.yml
```

The backend is organised as **vertical slices**: each folder under `features/`
holds every file needed to change one capability — its route, its request
decoding, its validation, its queries and its response mapping. Adding
`ExportBoard` means creating `features/boards/export.go` and registering it;
no other package needs to learn that the feature exists.

Two things deliberately stay horizontal:

- **`store/`** — sqlc emits one package, and keeping it whole preserves
  `DBTX` and `WithTx`. Slicing it per feature would make a cross-column card
  move impossible to express in one transaction.
- **`domain/`** — a Board is returned by five board slices and touched by nine
  column and card slices. Shared vocabulary has to live somewhere.

`platform/` holds the infrastructure the slices need — config, auth, the
database, and the HTTP plumbing. Nothing in `platform/` may mention Board,
Column or Card; if it does, that logic belongs in a slice.

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
make test           # Go unit tests (ordering, validation, dates, JWT)
make smoke          # every board/column/card endpoint, authenticated
make auth-smoke     # login, token handling, and cross-user isolation
make check          # fmt, vet, test, frontend build, frontend lint
cd frontend && npm run parity   # frontend ordering logic vs. the live API
```

`make smoke` and `make auth-smoke` need the stack running (`make up`). Both sign
in first and are safe to re-run: they clean up anything a previous interrupted
run left behind and assert that column and card positions stay a gap-free
`0..n-1` run after every move.

Point either at a different backend with `$env:KANBAN_API`:

```powershell
$env:KANBAN_API = 'http://localhost:8080/api/v1'
./backend/test/auth-smoke.ps1
```

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

**Access control.** Every board-scoped query filters on the authenticated user's
id, so there is no code path that reads a board without also constraining it to
its owner. Requesting someone else's board returns **404, not 403** — a 403
would confirm the board exists.

**`assignee` is free text, on purpose.** Boards are owner-only, so the only
person who can open a card is its owner. A user reference would name someone
who by construction cannot see the card. It names a person involved in the work
instead (a deal contact, a colleague with no account). See the comment in
`backend/sql/schema.sql` before "fixing" this.

**Dates.** `cards.due_date` is a calendar date and is served as `YYYY-MM-DD`.
It is never round-tripped through a timestamp, so it cannot shift by a day for
anyone west of UTC. The frontend formats it locally for display only.

**Sessions.** Stateless JWT in an `Authorization` header, held in
`localStorage`. Logout therefore cannot revoke a token — it stays valid until it
expires, which is why `JWT_TTL` defaults to 24h rather than "forever". Because
the credential is a header rather than a cookie, cross-site requests cannot
borrow it, so CSRF does not apply.

**The `date` type.** sqlc 1.31 does not honour a `date` override against the
pgx/v5 built-in mapping, so `due_date` stays `pgtype.Date` in the store. That
is deliberate: its `Valid` flag is exactly the nullable-calendar-date semantic
we want. See the note in `backend/sqlc.yaml`.

**Optimistic updates.** The frontend applies a drag immediately and reconciles
with the server afterwards, rolling back to the previous snapshot if the save
fails. See the frontend README for the details.

## API

Full reference in [`docs/API.md`](docs/API.md). In short — everything except
`/health` and `/auth/login` requires `Authorization: Bearer <token>`:

```
POST   /api/v1/auth/login
GET    /api/v1/auth/me

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
| `DATABASE_URL`         | `postgres://kanban:kanban@postgres:5432/kanban?sslmode=disable` |
| `BACKEND_PORT`         | `8080`                                           |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000`                          |
| `SEED_DEMO`            | `true` — seeds accounts only when the database has no users |
| `APP_ENV`              | `development`                                    |
| `JWT_SECRET`           | a built-in development default; **required** and ≥32 chars in production |
| `JWT_TTL`              | `24h`                                            |
| `DEMO_PASSWORD`        | `kanban-demo`                                    |

In development every `localhost` / `127.0.0.1` port is additionally allowed
through CORS, so the Next dev server can change ports freely. Set `APP_ENV` to
anything else and only the configured origins are accepted.

> **Postgres credentials.** `POSTGRES_USER` / `POSTGRES_PASSWORD` in
> docker-compose.yml are only read when Postgres first initialises an **empty**
> data directory. Changing them later has no effect on an existing `pgdata`
> volume — the container will start, report healthy, and reject the new
> credentials with no hint that the setting was ignored. Switching the current
> `kanban` user to anything else requires `docker compose down -v`, which
> destroys the database.

## Regenerating the store

After editing `backend/sql/queries.sql` or `backend/sql/schema.sql`:

```bash
make install-sqlc    # once
make generate
```

`make generate` fails loudly if a query and the schema have drifted apart —
that is the whole point of using sqlc instead of writing SQL by hand.
