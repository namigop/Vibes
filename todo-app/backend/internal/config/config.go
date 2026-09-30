package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	DatabaseURL        string
	Port               int
	Env                string
	LogLevel           string
	CORSAllowedOrigins []string
}

// Load reads env vars and validates the required ones.
// Defaults match .env.example so docker-compose works out of the box.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:        getenv("DATABASE_URL", "postgres://todo:todo@postgres:5432/todo?sslmode=disable"),
		Port:               getenvInt("BACKEND_PORT", 8080),
		Env:                getenv("APP_ENV", "development"),
		LogLevel:           getenv("LOG_LEVEL", "info"),
		CORSAllowedOrigins: splitCSV(getenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return nil, fmt.Errorf("BACKEND_PORT out of range: %d", cfg.Port)
	}
	if len(cfg.CORSAllowedOrigins) == 0 {
		return nil, errors.New("CORS_ALLOWED_ORIGINS must contain at least one origin")
	}
	return cfg, nil
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
