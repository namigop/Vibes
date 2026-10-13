package httpserver

import (
	"net/http"
	"net/url"
	"strings"
)

// corsMiddleware answers preflights and echoes a permitted origin.
//
// This API uses no cookies, so it never needs
// Access-Control-Allow-Credentials. In development every loopback port is
// accepted so the Next dev server can move between ports freely; outside
// development only the configured origins are allowed.
func (r *Router) corsMiddleware() func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(r.cfg.CORSAllowedOrigins))
	for _, o := range r.cfg.CORSAllowedOrigins {
		allowed[NormaliseOrigin(o)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			origin := req.Header.Get("Origin")

			if origin != "" && originAllowed(origin, allowed, r.cfg.IsDev()) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}

			if req.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", allowedRequestHeaders)
				w.Header().Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, req)
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

// NormaliseOrigin trims a trailing slash so a configured origin matches the
// Origin header exactly.
func NormaliseOrigin(raw string) string {
	return strings.TrimSuffix(strings.TrimSpace(raw), "/")
}
