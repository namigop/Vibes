package api

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"kanban/internal/config"
)

// NewRouter wires every route in the frozen contract (docs/API.md) onto the
// standard library's pattern-matching ServeMux, so there is no third-party
// router dependency to keep in step.
func NewRouter(svc *Service, cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	const base = "/api/v1"

	mux.HandleFunc("GET "+base+"/health", svc.handleHealth)

	mux.HandleFunc("GET "+base+"/boards", svc.handleListBoards)
	mux.HandleFunc("POST "+base+"/boards", svc.handleCreateBoard)
	mux.HandleFunc("GET "+base+"/boards/{boardId}", svc.handleGetBoard)
	mux.HandleFunc("PATCH "+base+"/boards/{boardId}", svc.handleRenameBoard)
	mux.HandleFunc("DELETE "+base+"/boards/{boardId}", svc.handleDeleteBoard)

	mux.HandleFunc("POST "+base+"/boards/{boardId}/columns", svc.handleCreateColumn)
	mux.HandleFunc("PATCH "+base+"/columns/{columnId}", svc.handleRenameColumn)
	mux.HandleFunc("POST "+base+"/columns/{columnId}/move", svc.handleMoveColumn)
	mux.HandleFunc("DELETE "+base+"/columns/{columnId}", svc.handleDeleteColumn)

	mux.HandleFunc("POST "+base+"/columns/{columnId}/cards", svc.handleCreateCard)
	mux.HandleFunc("PATCH "+base+"/cards/{cardId}", svc.handleUpdateCard)
	mux.HandleFunc("POST "+base+"/cards/{cardId}/move", svc.handleMoveCard)
	mux.HandleFunc("DELETE "+base+"/cards/{cardId}", svc.handleDeleteCard)

	// A bare /api/v1 hit is a misconfigured client; say so instead of 404ing
	// with the mux's unhelpful text.
	mux.HandleFunc("GET "+base+"/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, ErrorResponse{ErrNotFound, "unknown endpoint"})
	})

	var handler http.Handler = mux
	handler = corsMiddleware(cfg)(handler)
	handler = recoverMiddleware(handler)
	handler = logMiddleware(handler)

	return handler
}

// corsMiddleware answers preflights and echoes a permitted origin.
//
// This API uses no cookies or Authorization headers, so it never needs
// Access-Control-Allow-Credentials. In development every loopback port is
// accepted so the Next dev server can move between ports freely; in
// production only the configured origins are allowed.
func corsMiddleware(cfg config.Config) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(cfg.CORSAllowedOrigins))
	for _, o := range cfg.CORSAllowedOrigins {
		allowed[NormaliseOrigin(o)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" && originAllowed(origin, allowed, cfg.IsDev()) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func originAllowed(origin string, allowed map[string]struct{}, isDev bool) bool {
	if _, ok := allowed[origin]; ok {
		return true
	}
	if !isDev {
		return false
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := u.Hostname()
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// statusRecorder captures the response status for access logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
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
		if rec.status >= 500 {
			level = slog.LevelError
		} else if rec.status >= 400 {
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
					"method", r.Method,
					"path", r.URL.Path,
					"panic", rec,
				)
				writeJSON(w, http.StatusInternalServerError,
					ErrorResponse{ErrInternal, "an unexpected error occurred"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// NormaliseOrigin trims a trailing slash so configured origins match the
// Origin header exactly.
func NormaliseOrigin(raw string) string {
	return strings.TrimSuffix(strings.TrimSpace(raw), "/")
}
