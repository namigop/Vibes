package auth

import (
	"net/http"

	"kanban/internal/domain"
	"kanban/internal/platform/auth"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// me reports the account behind the bearer token.
//
// The client calls this on boot to discover whether a stored token is still
// valid; a 401 is how it learns it is signed out. That is why this route exists
// rather than the client decoding the token itself: only the server can verify
// the signature.
func me(q *store.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.CallerFrom(r.Context())
		if !ok {
			httpserver.WriteError(w, auth.ErrInvalidToken)
			return
		}

		user, err := q.GetUserByID(r.Context(), userID)
		if err != nil {
			if notFoundIfMissing(err, "user") != err {
				// The token names a user that no longer exists — the database was
				// reset underneath a still-valid token.
				httpserver.WriteError(w, auth.ErrInvalidToken)
				return
			}
			httpserver.WriteError(w, err)
			return
		}

		httpserver.WriteJSON(w, http.StatusOK, domain.UserFromStore(user))
	}
}
