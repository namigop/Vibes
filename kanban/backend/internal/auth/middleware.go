package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "kanban.userID"

// UserIDFrom returns the authenticated user's id, and whether one was present.
// Every protected handler depends on this being true, so the boolean is part of
// the signature rather than assumed.
func UserIDFrom(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

// MustUserID is for handlers that are only ever mounted behind RequireAuth,
// where a missing id would be a wiring bug rather than a client error.
func MustUserID(ctx context.Context) string {
	id, _ := UserIDFrom(ctx)
	return id
}

// withUserID stores the authenticated id on the request context.
func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
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

	credential = strings.TrimSpace(credential)
	return credential
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

			next.ServeHTTP(w, r.WithContext(withUserID(r.Context(), userID.String())))
		})
	}
}
