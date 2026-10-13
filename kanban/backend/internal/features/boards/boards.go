// Package boards contains the board slices: list, read, create, rename, delete.
//
// Each file owns one operation end to end — its route, its request body, its
// validation, its queries and its tests. Adding a board operation means adding
// a file to this folder, not editing a handler file, a service file and a types
// file that live in different places.
package boards

import (
	"net/http"

	"github.com/google/uuid"

	"kanban/internal/domain"
	"kanban/internal/platform/auth"
	"kanban/internal/platform/database"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// Register mounts every board route.
//
// This is the only file in the package that names an endpoint path, so the full
// surface of the feature is readable at a glance.
func Register(r *httpserver.Router, q *store.Queries) {
	r.Private("GET /api/v1/boards", listBoards(q))
	r.Private("GET /api/v1/boards/{boardId}", getBoard(q))
	r.Private("POST /api/v1/boards", createBoard(q))
	r.Private("PATCH /api/v1/boards/{boardId}", renameBoard(q))
	r.Private("DELETE /api/v1/boards/{boardId}", deleteBoard(q))
}

// caller resolves the authenticated user, writing a 401 and reporting false if
// there is none.
//
// Every slice needs it. It lives here so the pattern is written once, and so
// that a missing caller — which can only mean the route was mounted on Public
// by mistake — is loud rather than silent.
func caller(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, ok := auth.CallerFrom(r.Context())
	if !ok {
		httpserver.WriteError(w, auth.ErrInvalidToken)
		return uuid.Nil, false
	}
	return id, true
}

// notFoundIfMissing turns sqlc's ErrNoRows into a domain 404 and leaves every
// other error untouched, so a slice cannot accidentally turn a real database
// fault into a "not found".
func notFoundIfMissing(err error, what string) error {
	if database.IsNotFound(err) {
		return domain.NotFound("%s not found", what)
	}
	return err
}
