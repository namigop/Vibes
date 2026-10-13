package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"kanban/internal/domain"
	"kanban/internal/platform/auth"
)

// maxBodyBytes caps request bodies. A card is tiny; anything larger is a bug
// or an attack, and either way it should not be buffered.
const maxBodyBytes = 1 << 20 // 1 MiB

// allowedRequestHeaders is what a browser preflight may ask for.
//
// Authorization must be here. It is not a CORS-safelisted request header, so
// every authenticated browser request triggers a preflight that asks for it. If
// this list omits it, the browser blocks the real request before it is sent and
// fetch() surfaces that as a network error — indistinguishable from the server
// being down. curl, PowerShell and Go clients do not preflight, so no
// shell-based test can catch this.
const allowedRequestHeaders = "Authorization, Content-Type"

// ErrorResponse is the single error envelope every non-2xx response uses.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Stable machine-readable error slugs.
const (
	errValidation   = "validation_failed"
	errNotFound     = "not_found"
	errConflict     = "conflict"
	errUnauthorized = "unauthorized"
	errInternal     = "internal_error"
)

// Unauthorized builds the 401 a protected route returns when the bearer token
// is missing or unusable. It lives here so the router and the slices cannot
// disagree about the wording.
func Unauthorized() error {
	return domain.Unauthorized("authentication required: send 'Authorization: Bearer <token>'")
}

// WriteJSON renders v as the response body with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) {
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

// WriteError maps a domain error onto the envelope and an HTTP status.
//
// Translation lives here rather than in domain so that the domain package has
// no idea HTTP exists. It is centralised so no slice can invent its own status
// code for the same kind of failure.
func WriteError(w http.ResponseWriter, err error) {
	var (
		vErr domain.ValidationError
		nErr domain.NotFoundError
		cErr domain.ConflictError
		uErr domain.UnauthorizedError
	)

	// Tell the client this was an auth failure rather than a missing route.
	w.Header().Set("WWW-Authenticate", `Bearer realm="kanban"`)

	switch {
	case errors.As(err, &uErr):
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{errUnauthorized, uErr.Message})

	case errors.Is(err, auth.ErrInvalidCredentials):
		// Deliberately vague: confirming which half was wrong would let a
		// caller enumerate which usernames exist.
		WriteJSON(w, http.StatusUnauthorized,
			ErrorResponse{errUnauthorized, "invalid username or password"})

	case errors.Is(err, auth.ErrInvalidToken):
		WriteJSON(w, http.StatusUnauthorized,
			ErrorResponse{errUnauthorized, "session expired or invalid, please sign in again"})

	case errors.As(err, &vErr):
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{errValidation, vErr.Message})

	case errors.As(err, &nErr):
		WriteJSON(w, http.StatusNotFound, ErrorResponse{errNotFound, nErr.Message})

	case errors.As(err, &cErr):
		WriteJSON(w, http.StatusConflict, ErrorResponse{errConflict, cErr.Message})

	default:
		slog.Error("unhandled request error", "error", err)
		WriteJSON(w, http.StatusInternalServerError,
			ErrorResponse{errInternal, "an unexpected error occurred"})
	}
}

// DecodeJSON reads a JSON body into dst, rejecting malformed or oversized
// payloads with a validation error.
//
// Unknown fields are tolerated on purpose: an older client sending extra keys
// should not start failing against a newer server.
func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return domain.Invalid("request body is required")
	}

	dec := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes))
	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return domain.Invalid("request body is too large")
		}
		return domain.Invalid("request body must be valid JSON")
	}

	// Reject trailing content so `{"a":1}{"b":2}` is not silently accepted.
	if err := dec.Decode(new(struct{})); !errors.Is(err, io.EOF) {
		return domain.Invalid("request body must contain a single JSON object")
	}

	return nil
}

// PathUUID reads a path wildcard and validates it.
func PathUUID(r *http.Request, name, what string) (uuid.UUID, error) {
	return ParseUUID(r.PathValue(name), what)
}

// ParseUUID converts a string into a UUID or a domain validation error.
func ParseUUID(raw, what string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, domain.Invalid("%s is not a valid id", what)
	}
	return id, nil
}

// contextWithTimeout bounds handler work so a stuck query cannot pin a
// connection for the lifetime of the client. It is applied once by middleware,
// so handlers just use r.Context().
func contextWithTimeout(r *http.Request, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), d)
}

// statusRecorder captures the response status for access logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// logMiddleware emits one structured line per request.
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(rec, r)

		if rec.status == 0 {
			rec.status = http.StatusOK
		}

		level := slog.LevelInfo
		switch {
		case rec.status >= 500:
			level = slog.LevelError
		case rec.status >= 400:
			level = slog.LevelWarn
		}

		slog.Log(r.Context(), level, "http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"bytes", rec.bytes,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// recoverMiddleware turns a panic in a handler into a 500 instead of killing
// the process.
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic in handler",
					"method", r.Method, "path", r.URL.Path, "panic", rec)
				WriteJSON(w, http.StatusInternalServerError,
					ErrorResponse{errInternal, "an unexpected error occurred"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
