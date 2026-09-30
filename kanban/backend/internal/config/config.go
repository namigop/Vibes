// Package config loads runtime configuration from the environment.
//
// Every value has a development-friendly default so the server can be started
// with no setup at all, but DATABASE_URL is the one that really matters.
package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
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
}

// Addr is the listen address for the HTTP server.
func (c Config) Addr() string { return ":" + c.Port }

// IsDev reports whether the server is running in development mode.
func (c Config) IsDev() bool { return c.Env != "production" }

const defaultDatabaseURL = "postgres://todo:todo@localhost:5432/kanban?sslmode=disable"

// Load reads configuration from the environment, applying defaults.
func Load() Config {
	cfg := Config{
		DatabaseURL:        envOr("DATABASE_URL", defaultDatabaseURL),
		Port:               envOr("BACKEND_PORT", "8080"),
		CORSAllowedOrigins: splitAndTrim(envOr("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
		SeedDemo:           envBool("SEED_DEMO", true),
		Env:                envOr("APP_ENV", "development"),
	}
	return cfg
}

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
