.PHONY: help up down logs ps restart build test sqlc generate migrate-up migrate-down clean tidy

help:
	@echo Targets: up down logs ps restart build test sqlc generate migrate-up migrate-down clean tidy

up:
	docker compose up -d --build
	@echo Backend:  http://localhost:8080
	@echo Frontend: http://localhost:3000

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

restart:
	docker compose restart

build:
	docker compose build

# Run integration tests inside the backend container. Requires sqlc to have
# already generated internal/store/*.go (use `make generate` first).
test:
	docker compose run --rm backend go test ./...

# Generate sqlc code inside the backend container so we don't need sqlc on
# the host machine.
generate:
	docker compose run --rm backend sh -c "go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && sqlc generate"

# Direct sqlc invocation, expects sqlc on PATH (host machine).
sqlc:
	cd backend && sqlc generate

migrate-up:
	docker compose run --rm backend goose -dir /src/migrations postgres "$$DATABASE_URL" up

migrate-down:
	docker compose run --rm backend goose -dir /src/migrations postgres "$$DATABASE_URL" down

tidy:
	cd backend && go mod tidy

clean:
	docker compose down -v
