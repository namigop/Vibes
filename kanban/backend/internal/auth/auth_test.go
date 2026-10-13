package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if hash == "correct horse battery" {
		t.Fatal("the plaintext must never be returned")
	}
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("expected a bcrypt digest, got %q", hash)
	}
	if !CheckPassword(hash, "correct horse battery") {
		t.Error("the correct password should verify")
	}
	if CheckPassword(hash, "wrong") {
		t.Error("a wrong password must not verify")
	}
}

func TestHashIsSalted(t *testing.T) {
	a, _ := HashPassword("same-password")
	b, _ := HashPassword("same-password")
	if a == b {
		t.Error("two hashes of the same password should differ (random salt)")
	}
}

func TestValidatePasswordPolicy(t *testing.T) {
	if err := ValidatePasswordPolicy("short"); err == nil {
		t.Error("expected a minimum-length failure")
	}
	if err := ValidatePasswordPolicy("longenough"); err != nil {
		t.Errorf("unexpected rejection: %v", err)
	}
	if err := ValidatePasswordPolicy(strings.Repeat("a", 73)); err == nil {
		t.Error("expected rejection past bcrypt's 72-byte limit")
	}
	if err := ValidatePasswordPolicy("has\x00nul"); err == nil {
		t.Error("expected rejection of an embedded NUL")
	}
}

func TestSignerRoundTrip(t *testing.T) {
	signer := NewSigner("test-secret", time.Hour)
	userID := uuid.New()

	token, expiresAt, err := signer.Issue(userID, time.Now())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	if time.Until(expiresAt) > time.Hour+time.Minute {
		t.Errorf("expiry should be about an hour out, got %v", time.Until(expiresAt))
	}

	got, err := signer.Verify(token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got != userID {
		t.Errorf("Verify returned %v, want %v", got, userID)
	}
}

func TestSignerRejects(t *testing.T) {
	signer := NewSigner("test-secret", time.Hour)
	userID := uuid.New()

	t.Run("wrong secret", func(t *testing.T) {
		token, _, _ := signer.Issue(userID, time.Now())
		if _, err := NewSigner("other-secret", time.Hour).Verify(token); err == nil {
			t.Error("a token signed with another secret must not verify")
		}
	})

	t.Run("expired", func(t *testing.T) {
		// Issue far enough in the past that a one-minute TTL has elapsed.
		token, _, _ := signer.Issue(userID, time.Now().Add(-2*time.Hour))
		short := NewSigner("test-secret", time.Minute)
		if _, err := short.Verify(token); err == nil {
			t.Error("an expired token must not verify")
		}
	})

	t.Run("garbage", func(t *testing.T) {
		if _, err := signer.Verify("not-a-token"); err == nil {
			t.Error("expected rejection of a malformed token")
		}
	})

	t.Run("alg none", func(t *testing.T) {
		// The classic JWT bypass: an unsigned token claiming alg=none.
		unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{
			UserID: userID.String(),
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    tokenIssuer,
				Audience:  jwt.ClaimStrings{tokenAudience},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		})
		signed, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatalf("building the unsigned token failed: %v", err)
		}

		if _, err := signer.Verify(signed); err == nil {
			t.Error("an alg=none token must never be accepted")
		}
	})

	t.Run("malformed subject", func(t *testing.T) {
		// A correctly signed token whose subject is not a UUID must still fail.
		other := NewSigner("test-secret", time.Hour)
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
			UserID: "definitely-not-a-uuid",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    tokenIssuer,
				Audience:  jwt.ClaimStrings{tokenAudience},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
		}).SignedString([]byte("test-secret"))
		if err != nil {
			t.Fatalf("signing: %v", err)
		}

		if _, err := other.Verify(token); err == nil {
			t.Error("a token whose subject is not a UUID must be rejected")
		}
	})
}

func TestBearerFrom(t *testing.T) {
	cases := []struct {
		header string
		want   string
	}{
		{"Bearer abc123", "abc123"},
		{"bearer abc123", "abc123"},
		{"BEARER abc123", "abc123"},
		{"Bearer   abc123  ", "abc123"},
		{"", ""},
		{"Basic abc123", ""},
		{"abc123", ""},
		{"Bearer", ""},
		{"Bearer ", ""},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/boards", nil)
		if tc.header != "" {
			req.Header.Set("Authorization", tc.header)
		}

		if got := BearerFrom(req); got != tc.want {
			t.Errorf("BearerFrom(%q) = %q, want %q", tc.header, got, tc.want)
		}
	}
}

func TestRequireAuth(t *testing.T) {
	signer := NewSigner("test-secret", time.Hour)
	userID := uuid.New()
	valid, _, err := signer.Issue(userID, time.Now())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	cases := []struct {
		name     string
		header   string
		wantCode int
		wantUser string
	}{
		{"no header", "", http.StatusUnauthorized, ""},
		{"malformed header", "Basic abc", http.StatusUnauthorized, ""},
		{"bad token", "Bearer nonsense", http.StatusUnauthorized, ""},
		{"valid token", "Bearer " + valid, http.StatusOK, userID.String()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotUser string

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUser, _ = UserIDFrom(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			var code int
			reject := func(w http.ResponseWriter, r *http.Request) {
				code = http.StatusUnauthorized
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/boards", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}

			rec := httptest.NewRecorder()
			RequireAuth(signer, reject)(next).ServeHTTP(rec, req)

			// The pass-through handler sets 200 via WriteHeader; the reject
			// path never writes, so the recorder still holds its default 200
			// and `code` is what distinguishes them.
			if tc.wantCode == http.StatusUnauthorized {
				if code != http.StatusUnauthorized {
					t.Errorf("request should have been rejected, code=%d", code)
				}
				if gotUser != "" {
					t.Errorf("rejected request must carry no user, got %q", gotUser)
				}
				return
			}

			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", rec.Code)
			}
			if gotUser != tc.wantUser {
				t.Errorf("user = %q, want %q", gotUser, tc.wantUser)
			}
		})
	}
}
