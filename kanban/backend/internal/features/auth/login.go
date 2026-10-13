package auth

import (
	"net/http"
	"strings"
	"time"

	"kanban/internal/domain"
	"kanban/internal/platform/auth"
	"kanban/internal/platform/database"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse carries the bearer token and the signed-in user.
type LoginResponse struct {
	Token     string      `json:"token"`
	TokenType string      `json:"token_type"`
	ExpiresAt time.Time   `json:"expires_at"`
	User      domain.User `json:"user"`
}

// login verifies credentials and issues a token.
//
// An unknown username and a wrong password produce the same error and take
// comparable time, so the endpoint cannot be used to discover which accounts
// exist.
func login(q *store.Queries, s *auth.Signer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := httpserver.DecodeJSON(r, &req); err != nil {
			httpserver.WriteError(w, err)
			return
		}

		user, token, expiresAt, err := authenticate(r, q, s, req)
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		httpserver.WriteJSON(w, http.StatusOK, LoginResponse{
			Token:     token,
			TokenType: "Bearer",
			ExpiresAt: expiresAt,
			User:      user,
		})
	}
}

// authenticate holds the credential check and token issue so it is readable
// without the HTTP wrapper.
func authenticate(
	r *http.Request,
	q *store.Queries,
	s *auth.Signer,
	req LoginRequest,
) (domain.User, string, time.Time, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		return domain.User{}, "", time.Time{}, domain.Invalid("username and password are required")
	}

	user, err := q.GetUserByUsername(r.Context(), username)
	if err != nil {
		if database.IsNotFound(err) {
			// Still run a hash comparison so the response time does not reveal
			// whether the account exists.
			auth.CheckPassword(unknownUserHash, req.Password)
			return domain.User{}, "", time.Time{}, auth.ErrInvalidCredentials
		}
		return domain.User{}, "", time.Time{}, err
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		return domain.User{}, "", time.Time{}, auth.ErrInvalidCredentials
	}

	token, expiresAt, err := s.Issue(user.ID, time.Now().UTC())
	if err != nil {
		return domain.User{}, "", time.Time{}, err
	}

	return domain.UserFromStore(user), token, expiresAt, nil
}

// unknownUserHash is a valid bcrypt digest of a value nobody knows. Comparing
// against it keeps the unknown-user failure path as expensive as a real check,
// so response timing does not disclose whether the username exists.
const unknownUserHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
