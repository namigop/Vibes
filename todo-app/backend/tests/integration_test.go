package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver for goose

	"github.com/onezerotoys/todo-app/backend/internal/httpapi"
	"github.com/onezerotoys/todo-app/backend/internal/store"
	"github.com/onezerotoys/todo-app/backend/internal/todo"
)

// quietLogger keeps test output readable.
func quietLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

// setup spins up a Postgres 16 container, runs migrations, builds the router
// against the live DB and returns a cleanup func + a ready http.Handler.
func setup(t *testing.T) (http.Handler, func()) {
	t.Helper()
	ctx := context.Background()

	pgC, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		tcpostgres.WithDatabase("todo"),
		tcpostgres.WithUsername("todo"),
		tcpostgres.WithPassword("todo"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	dsn, err := pgC.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pgC.Terminate(ctx)
		t.Fatalf("dsn: %v", err)
	}

	// Run migrations against the container DB.
	stdDB, err := sql.Open("pgx", dsn)
	if err != nil {
		_ = pgC.Terminate(ctx)
		t.Fatalf("open sql db: %v", err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		stdDB.Close()
		_ = pgC.Terminate(ctx)
		t.Fatalf("goose dialect: %v", err)
	}
	migrationsDir := filepath.Join(projectRoot(t), "migrations")
	if err := goose.UpContext(ctx, stdDB, migrationsDir); err != nil {
		stdDB.Close()
		_ = pgC.Terminate(ctx)
		t.Fatalf("goose up: %v", err)
	}
	stdDB.Close()

	// Build the pgx pool the app actually uses.
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		_ = pgC.Terminate(ctx)
		t.Fatalf("parse cfg: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		_ = pgC.Terminate(ctx)
		t.Fatalf("open pool: %v", err)
	}

	queries := store.New(pool)
	svc := todo.NewService(queries)
	valid := todo.NewValidator()
	router := httpapi.NewRouter(svc, valid, quietLogger(), []string{"http://localhost:3000"})

	cleanup := func() {
		pool.Close()
		_ = pgC.Terminate(context.Background())
	}
	return router, cleanup
}

func TestTodosCRUDHappyPath(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	// Create
	createBody := `{"title":"buy milk","description":"2L semi-skimmed"}`
	rec := doRequest(t, router, http.MethodPost, "/api/v1/todos", createBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: want 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var created struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Done        bool   `json:"done"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.Title != "buy milk" || created.Description != "2L semi-skimmed" || created.Done {
		t.Fatalf("created mismatch: %+v", created)
	}
	if created.ID == "" {
		t.Fatalf("created id is empty")
	}

	// List (should contain exactly one)
	rec = doRequest(t, router, http.MethodGet, "/api/v1/todos", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list: want 200, got %d", rec.Code)
	}
	var listed struct {
		Items []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed.Items) != 1 || listed.Items[0].ID != created.ID {
		t.Fatalf("list mismatch: %+v", listed)
	}

	// Get
	rec = doRequest(t, router, http.MethodGet, "/api/v1/todos/"+created.ID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("get: want 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Patch: toggle done = true
	rec = doRequest(t, router, http.MethodPatch, "/api/v1/todos/"+created.ID, `{"done":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch: want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var patched struct {
		Done bool `json:"done"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &patched)
	if !patched.Done {
		t.Fatalf("patched.done should be true, got %+v", patched)
	}

	// Delete
	rec = doRequest(t, router, http.MethodDelete, "/api/v1/todos/"+created.ID, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: want 204, got %d", rec.Code)
	}

	// List again: empty
	rec = doRequest(t, router, http.MethodGet, "/api/v1/todos", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list: want 200, got %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed.Items) != 0 {
		t.Fatalf("list after delete: want 0 items, got %d", len(listed.Items))
	}
}

func TestValidationFailsOnEmptyTitle(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	rec := doRequest(t, router, http.MethodPost, "/api/v1/todos", `{"title":"","description":"x"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "validation_failed") {
		t.Fatalf("expected validation_failed code in body, got %s", rec.Body.String())
	}
}

func TestNotFoundOnUnknownID(t *testing.T) {
	router, cleanup := setup(t)
	defer cleanup()

	rec := doRequest(t, router, http.MethodGet, "/api/v1/todos/00000000-0000-0000-0000-000000000000", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "not_found") {
		t.Fatalf("expected not_found code in body, got %s", rec.Body.String())
	}
}

// doRequest is a small helper that buffers the request body and writes it.
func doRequest(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != "" {
		rdr = bytes.NewReader([]byte(body))
	}
	r := httptest.NewRequest(method, path, rdr)
	r.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

// projectRoot resolves the directory holding the migrations folder.
// `go test` invokes each package with cwd == the package directory, so for
// tests living in /backend/tests/ we walk up one level to /backend.
func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wd, "migrations")); err == nil {
		return wd
	}
	parent := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(parent, "migrations")); err == nil {
		return parent
	}
	// Last resort: scan upwards a few levels.
	dir := wd
	for i := 0; i < 5; i++ {
		dir = filepath.Dir(dir)
		if _, err := os.Stat(filepath.Join(dir, "migrations")); err == nil {
			return dir
		}
	}
	return wd
}
