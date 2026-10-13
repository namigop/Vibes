package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"kanban/internal/domain"
	"kanban/internal/platform/auth"
	"kanban/internal/platform/config"
)

// The Public/Private split is the router's one real invariant: a route is
// authenticated unless it deliberately opted out. Two ways to break it without
// noticing — forgetting to register a private route at all, or forgetting to
// mount the public group — both compile fine and only show up at runtime.

func newTestRouter(t *testing.T) *Router {
	t.Helper()
	cfg := config.Load()
	cfg.CORSAllowedOrigins = []string{"http://localhost:3000"}
	return New(cfg, auth.NewSigner("test-secret", time.Hour))
}

// tokenFor mints a valid bearer token so a private route can be reached.
func tokenFor(t *testing.T, signer *auth.Signer) string {
	t.Helper()
	token, _, err := signer.Issue(uuid.New(), time.Now())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	return token
}

// A Public route must actually be reachable without a token. This regressed
// once already: Handler() mounted only the private mux, so /health answered 401
// and the container healthcheck never passed.
func TestPublicRouteNeedsNoToken(t *testing.T) {
	router := newTestRouter(t)
	router.Public("GET /api/v1/health", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	rec := httptest.NewRecorder()
	router.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("public /health = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
}

// The flip side: anything not registered Public must be authenticated.
func TestPrivateRouteRequiresToken(t *testing.T) {
	router := newTestRouter(t)
	router.Private("GET /api/v1/boards", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{})
	})

	rec := httptest.NewRecorder()
	router.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/boards", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("private /boards without a token = %d, want 401", rec.Code)
	}
}

// Public routes are exact patterns and the private group is a "/api/v1/"
// catch-all. If the catch-all ever shadowed them, health checks would start
// requiring auth again — so assert both groups coexist.
func TestPublicAndPrivateCoexist(t *testing.T) {
	router := newTestRouter(t)
	router.Public("POST /api/v1/auth/login", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"token": "t"})
	})
	router.Private("GET /api/v1/auth/me", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"username": "demo"})
	})
	handler := router.Handler()

	t.Run("login is reachable without a token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("POST /auth/login = %d, want 200", rec.Code)
		}
	})

	t.Run("me still requires a token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("GET /auth/me = %d, want 401", rec.Code)
		}
	})
}

// docs/API.md promises every error response uses the envelope. A bare ServeMux
// answers with Go's plain-text 404, and answers 405 when the path exists under a
// different method — which reads as "wrong method on a real route" rather than
// "no such endpoint".
func TestUnknownEndpointUsesEnvelope(t *testing.T) {
	router := newTestRouter(t)
	router.Public("GET /api/v1/health", func(w http.ResponseWriter, _ *http.Request) {})
	// Only POST exists on this path.
	router.Private("POST /api/v1/boards/{id}/columns", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusCreated, map[string]string{})
	})
	handler := router.Handler()
	token := tokenFor(t, router.signer)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"no such path", http.MethodGet, "/api/v1/nope"},
		{"wrong method on a real path", http.MethodGet, "/api/v1/boards/abc/columns"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("%s %s = %d, want 404", tc.method, tc.path, rec.Code)
			}

			var body ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("body is not the error envelope: %v (%q)", err, rec.Body.String())
			}
			if body.Error != errNotFound {
				t.Errorf("error = %q, want %q", body.Error, errNotFound)
			}
			if body.Message == "" {
				t.Error("envelope must carry a message")
			}
		})
	}
}

// The envelope middleware must not swallow a legitimate 404 that a handler
// produced on purpose — a board that does not exist. That carries a real
// message, not "unknown endpoint".
func TestHandlerNotFoundKeepsItsMessage(t *testing.T) {
	router := newTestRouter(t)
	router.Private("GET /api/v1/boards/{id}", func(w http.ResponseWriter, _ *http.Request) {
		WriteError(w, domain.NotFound("board %s not found", uuid.New()))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/boards/"+uuid.New().String(), nil)
	req.Header.Set("Authorization", "Bearer "+tokenFor(t, router.signer))

	rec := httptest.NewRecorder()
	router.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var body ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not the envelope: %v (%q)", err, rec.Body.String())
	}
	if strings.Contains(body.Message, "unknown endpoint") {
		t.Errorf("a domain 404 was overwritten by the routing catch-all: %q", body.Message)
	}
}

// A successful response must survive the buffering middleware byte for byte,
// headers included. If the replay dropped Content-Type, every frontend fetch
// would start failing to parse a perfectly good response.
func TestSuccessfulResponsePassesThroughIntact(t *testing.T) {
	router := newTestRouter(t)
	router.Public("GET /api/v1/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Probe", "kept")
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	rec := httptest.NewRecorder()
	router.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	if got := rec.Header().Get("X-Probe"); got != "kept" {
		t.Errorf("X-Probe = %q, want %q (headers must survive the replay)", got, "kept")
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not valid JSON: %v (%q)", err, rec.Body.String())
	}
	if body["status"] != "ok" {
		t.Errorf("body = %v, want status ok", body)
	}
}
