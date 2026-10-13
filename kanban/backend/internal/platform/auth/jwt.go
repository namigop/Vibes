package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims is the JWT payload.
//
// It carries only what the API needs to authorise a request. Notably it does
// NOT carry the username: a rename must not invalidate outstanding tokens, and
// carrying display data would let a stale token show a name that no longer
// exists.
type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

// ErrInvalidToken covers every reason a token is unusable: bad signature,
// wrong issuer, expired, malformed. The caller must not surface which.
var ErrInvalidToken = errors.New("invalid or expired token")

const (
	tokenIssuer   = "kanban-api"
	tokenAudience = "kanban-web"
)

// Signer issues and verifies JWTs with a shared HMAC secret.
type Signer struct {
	secret []byte
	ttl    time.Duration
}

// NewSigner builds a Signer. A short TTL means less exposure from a token that
// leaked, at the cost of more frequent logins.
func NewSigner(secret string, ttl time.Duration) *Signer {
	return &Signer{secret: []byte(secret), ttl: ttl}
}

// Issue returns a signed token for a user plus its absolute expiry.
func (s *Signer) Issue(userID uuid.UUID, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(s.ttl)

	claims := Claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tokenIssuer,
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{tokenAudience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

// Verify validates a token and returns the user id it was issued for.
//
// The algorithm is pinned to HMAC. Without this, a token carrying `"alg":"none"`
// would be accepted as a valid, unsigned token — the classic JWT bypass.
func (s *Signer) Verify(tokenString string) (uuid.UUID, error) {
	var claims Claims

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(t *jwt.Token) (any, error) { return s.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithAudience(tokenAudience),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	return userID, nil
}

// TTL is how long issued tokens remain valid.
func (s *Signer) TTL() time.Duration { return s.ttl }
