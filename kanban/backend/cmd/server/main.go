// Command server runs the kanban HTTP API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kanban/internal/api"
	"kanban/internal/config"
	"kanban/internal/store"
	"kanban/migrations"
)

const (
	// shutdownGrace is how long in-flight requests get to finish.
	shutdownGrace = 10 * time.Second
	readTimeout   = 15 * time.Second
	writeTimeout  = 30 * time.Second
	idleTimeout   = 60 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	setupLogging(cfg)

	slog.Info("starting kanban api",
		"addr", cfg.Addr(),
		"env", cfg.Env,
		"database", cfg.Redacted(),
		"seed_demo", cfg.SeedDemo,
	)

	// bootCtx is cancelled by Ctrl-C; everything before the server starts
	// uses it so a shutdown during startup unwinds cleanly.
	bootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := store.Open(bootCtx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	slog.Info("connected to database")

	if err := migrations.Run(bootCtx, pool); err != nil {
		return err
	}
	slog.Info("migrations up to date")

	if cfg.SeedDemo {
		if err := store.SeedDemoData(bootCtx, pool); err != nil {
			return err
		}
		slog.Info("demo data ensured")
	}

	svc := api.NewService(pool)
	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           api.NewRouter(svc, cfg),
		ReadHeaderTimeout: readTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		return err
	case <-bootCtx.Done():
		slog.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	slog.Info("stopped cleanly")
	return nil
}

func setupLogging(cfg config.Config) {
	level := slog.LevelDebug
	if !cfg.IsDev() {
		level = slog.LevelInfo
	}

	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	if cfg.IsDev() {
		// Human-readable logs are much easier to scan during development.
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(handler))
}
