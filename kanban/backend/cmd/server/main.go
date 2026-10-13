// Command server is the composition root.
//
// Its only job is to assemble the application: read config, open the database,
// apply migrations, seed demo data if enabled, register every feature slice, and
// listen. There is no business logic here — that belongs to internal/features,
// one slice per operation.
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

	"github.com/jackc/pgx/v5/pgxpool"

	"kanban/internal/devseed"
	"kanban/internal/features/auth"
	"kanban/internal/features/boards"
	"kanban/internal/features/cards"
	"kanban/internal/features/columns"
	platformauth "kanban/internal/platform/auth"
	"kanban/internal/platform/config"
	"kanban/internal/platform/database"
	"kanban/internal/platform/database/migrations"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
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

	// Refuse to boot a configuration that is only safe on a laptop. Seeded
	// demo personas share one published password and there is no signup flow,
	// so exposing that to a network is a full account compromise.
	if err := cfg.Validate(); err != nil {
		return err
	}
	if cfg.UsesDefaultJWTSecret() {
		slog.Warn("using the built-in development JWT secret; " +
			"anyone can mint tokens for any user. Set JWT_SECRET before exposing this.")
	}

	signer := platformauth.NewSigner(cfg.JWTSecret, cfg.JWTTTL)

	slog.Info("starting kanban api",
		"addr", cfg.Addr(),
		"env", cfg.Env,
		"database", cfg.Redacted(),
		"seed_demo", cfg.SeedDemo,
		"token_ttl", cfg.JWTTTL,
	)

	// bootCtx is cancelled by Ctrl-C; everything before the server starts uses
	// it so a shutdown during startup unwinds cleanly.
	bootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Open(bootCtx, cfg.DatabaseURL)
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
		if err := seedDemo(bootCtx, pool, cfg); err != nil {
			return err
		}
	}

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           buildHandler(cfg, pool, signer),
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

// buildHandler assembles the router from the feature slices.
//
// This list is the whole application surface in one place: adding a feature
// means adding a package under internal/features and one line here. Nothing
// else in the codebase needs to know it exists.
func buildHandler(cfg config.Config, pool *pgxpool.Pool, signer *platformauth.Signer) http.Handler {
	queries := store.New(pool)
	router := httpserver.New(cfg, signer)

	auth.Register(router, queries, pool, signer)
	boards.Register(router, queries)
	columns.Register(router, queries, pool)
	cards.Register(router, queries, pool)

	return router.Handler()
}

// seedDemo provisions demo accounts, including `legacy` so its password tracks
// DEMO_PASSWORD too. legacy owns boards that predate ownership and is given no
// demo content — it is the archive account, not a persona.
func seedDemo(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	personas := append(append([]string{}, config.DemoPersonas...), config.LegacyUsername)

	seeded, err := devseed.Run(ctx, pool, personas, cfg.DemoPassword)
	if err != nil {
		return err
	}

	if len(seeded.CreatedPersonas) == 0 {
		slog.Info("demo accounts already present, nothing seeded")
	} else {
		slog.Info("seeded demo accounts",
			"personas", seeded.CreatedPersonas,
			"boards", seeded.CreatedBoards)
	}
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
