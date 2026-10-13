package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const callerKey contextKey = "kanban.caller"

// CallerFrom returns the authenticated user id and whether one was present.
//
// Slices depend on this being present, so the boolean is part of the signature
// rather than assumed.
func CallerFrom(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(callerKey).(uuid.UUID)
	return id, ok
}

// MustCaller is for code that only runs behind RequireAuth, where a missing
// caller would be a wiring bug rather than a client error.
func MustCaller(ctx context.Context) uuid.UUID {
	id, _ := CallerFrom(ctx)
	return id
}

func withCaller(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, callerKey, id)
}

// BearerFrom extracts the credential from an Authorization header.
//
// It accepts only the `Bearer <token>` form, case-insensitively per RFC 7235,
// and returns "" when the header is absent or malformed.
func BearerFrom(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}

	scheme, credential, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}

	return strings.TrimSpace(credential)
}

// RequireAuth rejects any request without a valid bearer token.
//
// There is no session store: a token is trusted purely on its signature and
// expiry. The consequence is that logout cannot revoke an already-issued
// token — it stays valid until it expires. That is the deliberate trade for a
// stateless API, bounded by the short TTL.
func RequireAuth(signer *Signer, onUnauthorized http.HandlerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			credential := BearerFrom(r)
			if credential == "" {
				onUnauthorized(w, r)
				return
			}

			userID, err := signer.Verify(credential)
			if err != nil {
				onUnauthorized(w, r)
				return
			}

			next.ServeHTTP(w, r.WithContext(withCaller(r.Context(), userID)))
		})
	}
}
