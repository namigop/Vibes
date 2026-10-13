package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kanban/internal/platform/config"
)

// newCORSRouter builds a router whose only job in these tests is the CORS
// middleware. The bearer signer is nil because no route is ever reached: a
// preflight is answered by the middleware before routing.
func newCORSRouter(t *testing.T, cfg config.Config) http.Handler {
	t.Helper()
	return New(cfg, nil).corsMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}

// A browser sends a preflight before any request carrying `Authorization`,
// because that header is not CORS-safelisted. If the preflight does not
// advertise it, the browser blocks the real request before it leaves the page
// and fetch() reports a network error — which is indistinguishable from "the
// backend is down".
//
// Non-browser clients (curl, PowerShell, Go) do not preflight, so no
// end-to-end test against the running server can catch this. It has to be
// asserted here, against the middleware.
func TestPreflightAllowsAuthorizationHeader(t *testing.T) {
	cfg := config.Load()
	cfg.CORSAllowedOrigins = []string{"http://localhost:3000"}

	router := New(cfg, nil)
	handler := router.corsMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("a preflight must not reach the wrapped handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/boards", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "authorization")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}

	allowHeaders := rec.Header().Get("Access-Control-Allow-Headers")
	if !strings.Contains(strings.ToLower(allowHeaders), "authorization") {
		t.Errorf(
			"Access-Control-Allow-Headers = %q; it must include \"authorization\", "+
				"otherwise every authenticated browser request is blocked by the preflight",
			allowHeaders,
		)
	}

	// Content-Type must still be allowed: bodies are JSON.
	if !strings.Contains(strings.ToLower(allowHeaders), "content-type") {
		t.Errorf("Access-Control-Allow-Headers = %q; must still include \"content-type\"", allowHeaders)
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Access-Control-Allow-Origin = %q, want http://localhost:3000", got)
	}
}

func TestPreflightAllowsAllMethodsUsedByTheClient(t *testing.T) {
	cfg := config.Load()
	cfg.CORSAllowedOrigins = []string{"http://localhost:3000"}

	handler := newCORSRouter(t, cfg)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/boards", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	allowMethods := rec.Header().Get("Access-Control-Allow-Methods")
	for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
		if !strings.Contains(allowMethods, method) {
			t.Errorf("Access-Control-Allow-Methods = %q; missing %s", allowMethods, method)
		}
	}
}

func TestCORSRejectsUnknownOriginInProduction(t *testing.T) {
	cfg := config.Load()
	cfg.Env = "production"
	cfg.CORSAllowedOrigins = []string{"https://kanban.example.com"}

	handler := newCORSRouter(t, cfg)

	t.Run("known origin is echoed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/boards", nil)
		req.Header.Set("Origin", "https://kanban.example.com")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got == "" {
			t.Error("a configured origin should be echoed back")
		}
	})

	t.Run("unknown origin gets no allow-origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/boards", nil)
		req.Header.Set("Origin", "http://evil.example.com")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("Access-Control-Allow-Origin = %q, want empty for a disallowed origin", got)
		}
	})
}

func TestOriginAllowedLoopbackOnlyInDev(t *testing.T) {
	allowed := map[string]struct{}{NormaliseOrigin("http://localhost:3000"): {}}

	if !originAllowed("http://localhost:5173", allowed, true) {
		t.Error("any loopback port should be allowed in development")
	}
	if originAllowed("http://localhost:5173", allowed, false) {
		t.Error("loopback should NOT be auto-allowed outside development")
	}
	if originAllowed("http://evil.example.com", allowed, true) {
		t.Error("a remote origin must never be allowed, even in development")
	}
}
