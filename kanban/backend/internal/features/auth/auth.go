// Package auth contains the authentication slices: exchanging credentials for
// a token, and reporting who the caller is.
//
// Login is the only data route in the app registered as Public; everything else
// in the codebase is Private. This package is where that boundary is visible.
package auth

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"kanban/internal/domain"
	"kanban/internal/platform/auth"
	"kanban/internal/platform/database"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// Register mounts the auth routes.
func Register(r *httpserver.Router, q *store.Queries, pool *pgxpool.Pool, s *auth.Signer) {
	r.Public("GET /api/v1/health", health(pool))
	r.Public("POST /api/v1/auth/login", login(q, s))
	r.Private("GET /api/v1/auth/me", me(q))
}

// health reports API and database liveness.
//
// It is public because orchestrators and load balancers poll it, and because an
// unauthenticated caller should be able to tell "the API is down" apart from
// "your token expired".
func health(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			httpserver.WriteJSON(w, http.StatusServiceUnavailable,
				domain.Health{Status: "degraded", DB: "unreachable"})
			return
		}
		httpserver.WriteJSON(w, http.StatusOK, domain.Health{Status: "ok", DB: "ok"})
	}
}

// notFoundIfMissing turns sqlc's ErrNoRows into a domain 404.
func notFoundIfMissing(err error, what string) error {
	if database.IsNotFound(err) {
		return domain.NotFound("%s not found", what)
	}
	return err
}
