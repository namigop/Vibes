# Kanban — Go + Postgres + Next.js
#
# Windows note: run these through Git Bash, WSL, or `make` from a POSIX shell.
# The commands below assume a POSIX shell because the Go toolchain and npm
# scripts are invoked the same way on every platform.

BACKEND := backend
FRONTEND := frontend
SQLC := sqlc
GO := go

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# --- database ---------------------------------------------------------------

.PHONY: db-up
db-up: ## Start just Postgres in Docker
	docker compose up -d postgres

.PHONY: db-down
db-down: ## Stop and remove the Postgres container (keeps the volume)
	docker compose down

.PHONY: db-reset
db-reset: ## Destroy the Postgres volume and start clean
	docker compose down -v
	docker compose up -d postgres

# --- code generation --------------------------------------------------------

.PHONY: generate
generate: ## Regenerate the sqlc store from sql/schema.sql + sql/queries.sql
	cd $(BACKEND) && $(SQLC) generate

.PHONY: install-sqlc
install-sqlc: ## Install the sqlc code generator
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# --- backend ----------------------------------------------------------------

.PHONY: build
build: ## Build the API binary into backend/bin
	cd $(BACKEND) && $(GO) build -o bin/server ./cmd/server

.PHONY: test
test: ## Run Go unit tests
	cd $(BACKEND) && $(GO) test ./...

.PHONY: vet
vet: ## Run go vet
	cd $(BACKEND) && $(GO) vet ./...

.PHONY: fmt
fmt: ## Format Go sources
	cd $(BACKEND) && gofmt -w ./cmd ./internal ./migrations

.PHONY: api
api: ## Run the API in the foreground
	cd $(BACKEND) && $(GO) run ./cmd/server

.PHONY: smoke
smoke: ## Exercise every endpoint against a running API
	./$(BACKEND)/test/smoke.ps1

# --- frontend ---------------------------------------------------------------

.PHONY: web
web: ## Run the Next.js dev server
	cd $(FRONTEND) && npm run dev

.PHONY: web-build
web-build: ## Production build of the frontend
	cd $(FRONTEND) && npm run build

.PHONY: lint
lint: ## Lint the frontend
	cd $(FRONTEND) && npm run lint

# --- everything -------------------------------------------------------------

.PHONY: check
check: fmt vet test web-build lint ## Everything CI would run

.PHONY: up
up: ## Run the full stack in Docker
	docker compose up --build
