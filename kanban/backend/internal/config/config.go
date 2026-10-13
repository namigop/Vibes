// Package config loads runtime configuration from the environment.
//
// Every value has a development-friendly default so the server can be started
// with no setup at all, but DATABASE_URL is the one that really matters.
package config

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the fully-resolved runtime configuration.
type Config struct {
	DatabaseURL string
	Port        string
	// CORSAllowedOrigins is a comma-separated list of exact origins that are
	// always allowed. In development every localhost/127.0.0.1 port is also
	// allowed (see api.NewCORSHandler) so the Next.js dev server can move
	// between ports without reconfiguring the API.
	CORSAllowedOrigins []string
	SeedDemo           bool
	Env                string

	// JWTSecret signs and verifies bearer tokens. It must be set explicitly in
	// production; in development a fixed default is used and logged loudly,
	// because a real deployment with a default secret would let anyone mint a
	// token for any user.
	JWTSecret string
	// JWTTTL is how long an issued token stays valid. It is also the window in
	// which a leaked token remains usable, since tokens are stateless and
	// cannot be revoked.
	JWTTTL time.Duration
	// DemoPassword is the shared password given to every seeded persona.
	// Seeded accounts are only ever created in development (see Validate).
	DemoPassword string
}

// Addr is the listen address for the HTTP server.
func (c Config) Addr() string { return ":" + c.Port }

// IsDev reports whether the server is running in development mode.
func (c Config) IsDev() bool { return c.Env != "production" }

// defaultDatabaseURL is the in-network address: the `postgres` service name
// and its container port 5432. It only resolves from inside docker compose.
//
// When running the backend directly on the host against the compose database,
// override it with the *published* port instead:
//
//	DATABASE_URL=postgres://kanban:kanban@localhost:5433/kanban?sslmode=disable
//
// Note the absence of a trailing space — a DSN with one will not parse.
//
// The kanban/kanban credentials here match the existing `pgdata` volume.
// Postgres applies POSTGRES_USER / POSTGRES_PASSWORD only when it first
// initialises an empty data directory, so changing them in docker-compose.yml
// has no effect on a volume that already exists. Switching to another username
// requires `docker compose down -v`, which destroys the database.
const defaultDatabaseURL = "postgres://kanban:kanban@postgres:5432/kanban?sslmode=disable"

// Load reads configuration from the environment, applying defaults.
func Load() Config {
	cfg := Config{
		DatabaseURL:        envOr("DATABASE_URL", defaultDatabaseURL),
		Port:               envOr("BACKEND_PORT", "8080"),
		CORSAllowedOrigins: splitAndTrim(envOr("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
		SeedDemo:           envBool("SEED_DEMO", true),
		Env:                envOr("APP_ENV", "development"),
		JWTSecret:          envOr("JWT_SECRET", defaultJWTSecret),
		JWTTTL:             envDuration("JWT_TTL", defaultJWTTTL),
		DemoPassword:       envOr("DEMO_PASSWORD", defaultDemoPassword),
	}
	return cfg
}

// Validate rejects configurations that are safe in development but not in
// production.
//
// The important one is seeded demo personas. There is no signup flow, so every
// account comes from the seed and every account shares DEMO_PASSWORD — anyone
// who can reach a production instance could log in as anyone. Rather than
// trusting a convention, the server refuses to boot.
func (c Config) Validate() error {
	if c.IsDev() {
		return nil
	}

	if c.SeedDemo {
		return errors.New(
			"APP_ENV=production but SEED_DEMO is enabled: demo personas share a " +
				"known password and there is no signup flow, so anyone who can " +
				"reach this server could log in as any user. Set SEED_DEMO=false " +
				"and provision real accounts.")
	}

	if c.JWTSecret == defaultJWTSecret {
		return errors.New(
			"APP_ENV=production but JWT_SECRET is still the built-in development " +
				"default. Anyone could mint a valid token for any user id.")
	}

	if len(c.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters in production")
	}

	if c.DemoPassword == defaultDemoPassword {
		return errors.New(
			"APP_ENV=production but DEMO_PASSWORD is still the default. Any " +
				"seeded account would be reachable with a published password.")
	}

	return nil
}

// UsesDefaultJWTSecret reports whether the server is signing with the
// development fallback. The caller logs this as a warning at boot.
func (c Config) UsesDefaultJWTSecret() bool { return c.JWTSecret == defaultJWTSecret }

const (
	// defaultJWTSecret is intentionally obvious. It exists so `go run` works
	// with no setup, and Validate() refuses to let it reach production.
	defaultJWTSecret = "dev-only-insecure-jwt-secret-change-me"

	// defaultJWTTTL bounds how long a leaked token stays useful. Tokens are
	// stateless, so this is the only limit on their lifetime.
	defaultJWTTTL = 24 * time.Hour

	// defaultDemoPassword is shared by every seeded persona.
	defaultDemoPassword = "kanban-demo"
)

// DemoPersonas are the accounts the seed provisions. There is no signup route,
// so this list is the entire user population of a fresh install.
var DemoPersonas = []string{"alice", "bob", "carol", "demo"}

// LegacyUsername owns boards that predate ownership.
const LegacyUsername = "legacy"

// Redacted returns a copy safe to log: the password in DATABASE_URL is masked.
func (c Config) Redacted() slog.Value {
	return slog.StringValue(RedactDatabaseURL(c.DatabaseURL))
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	b, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return fallback
	}
	return b
}

// envDuration parses a Go duration string such as "24h" or "30m", falling back
// when unset or unparseable rather than failing to boot.
func envDuration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	d, err := time.ParseDuration(strings.TrimSpace(v))
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
