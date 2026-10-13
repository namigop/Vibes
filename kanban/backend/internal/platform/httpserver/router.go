// Package httpserver holds the HTTP plumbing shared by every feature slice:
// routing, CORS, the JSON error envelope, body decoding, and the cross-cutting
// middleware.
//
// The rule for this package is narrow: no business rules live here, and nothing
// in it may mention Board, Column or Card.
//
// It does import internal/domain, but only to translate the four domain error
// kinds into status codes in WriteError. That coupling is deliberate and
// one-directional: domain has no idea HTTP exists, and a slice cannot invent its
// own status code for a kind of failure. If you find yourself needing to know
// what a Board is in order to change this package, the logic belongs in a
// feature slice instead.
package httpserver

import (
	"net/http"
	"time"

	"kanban/internal/platform/auth"
	"kanban/internal/platform/config"
)

// requestTimeout bounds every request. It is applied once, as middleware, so a
// handler can simply use r.Context() instead of deriving and cancelling a
// context of its own.
const requestTimeout = 15 * time.Second

// Router collects routes from feature slices and serves them.
//
// There are two groups rather than one, and the split is the point: a slice
// registers against Public or Private, and Private is wrapped in auth
// middleware as a unit. A new endpoint is therefore authenticated by default,
// because forgetting middleware is not something you can forget to do — only
// the deliberate choice of Public makes a route unauthenticated.
type Router struct {
	cfg    config.Config
	signer *auth.Signer

	private *http.ServeMux

	// publicRoutes is recorded rather than kept in a mux of its own. A second
	// ServeMux would not fall through: ServeMux picks the single most specific
	// matching pattern and stops, so an unmatched public path would never reach
	// the private group. Keeping the registrations lets Handler mount them as
	// exact patterns, which are more specific than the "/api/v1/" catch-all and
	// therefore win.
	publicRoutes []route
}

type route struct {
	pattern string
	handler http.HandlerFunc
}

// New creates an empty router.
func New(cfg config.Config, signer *auth.Signer) *Router {
	r := &Router{
		cfg:     cfg,
		signer:  signer,
		private: http.NewServeMux(),
	}
	// A method-less catch-all for anything under /api/v1/. Registering it here,
	// rather than rewriting 404s afterwards, is what makes the contract hold:
	//
	//   - docs/API.md promises every error uses the envelope, and a bare ServeMux
	//     answers with Go's plain-text 404 page.
	//   - It also prevents a 405. Without a method-less match, ServeMux answers
	//     405 when the path exists under a different method, which reads as
	//     "wrong method on a real route" rather than "no such endpoint".
	//   - And it leaves a handler's own 404 alone, because a matched handler is
	//     always more specific than this catch-all.
	r.private.HandleFunc("/api/v1/", unknownEndpoint)
	return r
}

// unknownEndpoint is the envelope 404 for a path no slice registered.
func unknownEndpoint(w http.ResponseWriter, _ *http.Request) {
	WriteJSON(w, http.StatusNotFound, ErrorResponse{errNotFound, "unknown endpoint"})
}

// Public registers an unauthenticated route.
//
// Only genuinely public endpoints belong here: /health and /auth/login.
// Anything data-bearing should use Private so that forgetting auth is loud
// rather than silent.
func (r *Router) Public(pattern string, handler http.HandlerFunc) {
	r.publicRoutes = append(r.publicRoutes, route{pattern, handler})
}

// Private registers a route behind bearer-token authentication.
func (r *Router) Private(pattern string, handler http.HandlerFunc) {
	r.private.HandleFunc(pattern, handler)
}

// Handler wires the two groups together with the cross-cutting middleware and
// returns the final http.Handler.
func (r *Router) Handler() http.Handler {
	onUnauthorized := func(w http.ResponseWriter, _ *http.Request) {
		WriteError(w, Unauthorized())
	}

	mux := http.NewServeMux()
	// Public first. These are exact patterns, so ServeMux prefers them over the
	// "/api/v1/" catch-all mounted below.
	for _, pub := range r.publicRoutes {
		mux.HandleFunc(pub.pattern, pub.handler)
	}
	// Everything else is authenticated by default.
	mux.Handle("/api/v1/", auth.RequireAuth(r.signer, onUnauthorized)(r.private))
	// Outside the API namespace entirely.
	mux.HandleFunc("/", unknownEndpoint)

	var handler http.Handler = mux
	handler = r.corsMiddleware()(handler)
	handler = withRequestTimeout(requestTimeout)(handler)
	handler = recoverMiddleware(handler)
	handler = logMiddleware(handler)

	return handler
}

// withRequestTimeout bounds handler work so a stuck query cannot pin a
// connection for the lifetime of the client.
func withRequestTimeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := contextWithTimeout(r, d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
