package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ValidationError is raised for bad client input and rendered as HTTP 400.
type ValidationError struct{ Message string }

func (e ValidationError) Error() string { return e.Message }

// NotFoundError is raised when an addressed resource does not exist and is
// rendered as HTTP 404.
type NotFoundError struct{ Message string }

func (e NotFoundError) Error() string { return e.Message }

// ConflictError is rendered as HTTP 409.
type ConflictError struct{ Message string }

func (e ConflictError) Error() string { return e.Message }

func invalid(format string, args ...any) error {
	return ValidationError{Message: sprintf(format, args...)}
}

func notFound(format string, args ...any) error {
	return NotFoundError{Message: sprintf(format, args...)}
}

// writeJSON serialises v as the response body with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// The status line is already sent, so this can only be logged.
		slog.Error("write response", "error", err)
	}
}

// writeError maps a domain error onto the frozen error envelope.
//
// pgx.ErrNoRows is translated to 404 here so no handler has to remember to do
// it — an UPDATE ... RETURNING with no match and a SELECT miss are the same
// thing to a client.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		vErr ValidationError
		nErr NotFoundError
		cErr ConflictError
	)

	switch {
	case errors.As(err, &vErr):
		writeJSON(w, http.StatusBadRequest, ErrorResponse{ErrValidation, vErr.Message})
	case errors.As(err, &nErr):
		writeJSON(w, http.StatusNotFound, ErrorResponse{ErrNotFound, nErr.Message})
	case errors.As(err, &cErr):
		writeJSON(w, http.StatusConflict, ErrorResponse{ErrConflict, cErr.Message})
	case errors.Is(err, pgx.ErrNoRows):
		writeJSON(w, http.StatusNotFound, ErrorResponse{ErrNotFound, "resource not found"})
	default:
		slog.Error("unhandled request error",
			"method", r.Method, "path", r.URL.Path, "error", err)
		writeJSON(w, http.StatusInternalServerError,
			ErrorResponse{ErrInternal, "an unexpected error occurred"})
	}
}

// ---- input validation helpers ---------------------------------------------

// parseUUID converts a path segment into a UUID or a 400.
func parseUUID(raw, what string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, invalid("%s is not a valid id", what)
	}
	return id, nil
}

// requireName trims and validates a human-supplied name.
func requireName(raw, what string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", invalid("%s must not be empty", what)
	}
	if len(name) > 120 {
		return "", invalid("%s must be 120 characters or fewer", what)
	}
	return name, nil
}

// clampIndex keeps a destination index inside [0, n] so a client that is a
// little out of date (or malicious) can never corrupt the ordering.
func clampIndex(pos int32, n int32) int32 {
	if pos < 0 {
		return 0
	}
	if pos > n {
		return n
	}
	return pos
}
