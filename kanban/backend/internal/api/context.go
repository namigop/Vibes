package api

import (
	"context"
	"net/http"
	"time"
)

// ctxT is an alias kept local so handlers can express their return type
// without importing context into every file.
type ctxT = context.Context

// withTimeout derives a request-scoped context with a deadline.
func withTimeout(r *http.Request, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), d)
}
