package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver for goose
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	"github.com/onezerotoys/todo-app/backend/internal/config"
	"github.com/onezerotoys/todo-app/backend/internal/httpapi"
	"github.com/onezerotoys/todo-app/backend/internal/store"
	"github.com/onezerotoys/todo-app/backend/internal/todo"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config load failed", slog.Any("err", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := openPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db pool open failed", slog.Any("err", err))
		os.Exit(1)
	}
	defer pool.Close()

	if err := runMigrations(ctx, cfg.DatabaseURL, "migrations"); err != nil {
		logger.Error("migrations failed", slog.Any("err", err))
		os.Exit(1)
	}

	queries := store.New(pool)
	svc := todo.NewService(queries)
	valid := todo.NewValidator()

	router := httpapi.NewRouter(svc, valid, logger, cfg.CORSAllowedOrigins)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in background
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server listening", slog.String("addr", srv.Addr), slog.String("env", cfg.Env))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
	case err := <-serverErr:
		logger.Error("server error", slog.Any("err", err))
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.Any("err", err))
		os.Exit(1)
	}
	logger.Info("server stopped")
}

func openPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return pool, nil
}

// runMigrations runs all pending goose migrations from the given directory.
// We use the stdlib database/sql-compatible interface by opening a parallel
// sql.DB with the same DSN; this keeps goose happy without a separate CLI run.
func runMigrations(ctx context.Context, dsn, dir string) error {
	stdDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open sql db: %w", err)
	}
	defer stdDB.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := stdDB.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.UpContext(ctx, stdDB, dir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
