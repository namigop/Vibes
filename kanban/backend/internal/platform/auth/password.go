// Package auth implements password hashing, JWT issuing/verification, and the
// bearer-token middleware that every protected endpoint passes through.
package auth

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// minPasswordLength is the shortest password we will store. This is not a
// substitute for entropy, but it rejects the obvious mistakes.
const minPasswordLength = 8

// bcryptMaxLength is bcrypt's own limit: anything past 72 bytes is silently
// ignored by the algorithm, so a longer password would be no stronger than its
// first 72 bytes. Rejecting it is clearer than pretending it counts.
const bcryptMaxLength = 72

// ErrInvalidCredentials is returned for both an unknown username and a wrong
// password. Callers must not distinguish them — doing so leaks which usernames
// exist.
var ErrInvalidCredentials = errors.New("invalid username or password")

// HashPassword produces a bcrypt digest at the default cost.
func HashPassword(plain string) (string, error) {
	digest, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(digest), nil
}

// CheckPassword reports whether plain matches the stored bcrypt digest.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// ValidatePasswordPolicy enforces the minimum length rules.
//
// bcrypt is deliberately not in the list of allowed characters: a password
// containing a NUL byte cannot be hashed correctly, and silently truncating
// it would produce a hash that accepts a shorter password than intended.
func ValidatePasswordPolicy(plain string) error {
	if len(plain) < minPasswordLength {
		return errors.New("password must be at least 8 characters")
	}
	if len(plain) > bcryptMaxLength {
		return errors.New("password must be 72 bytes or fewer")
	}
	if strings.ContainsRune(plain, 0) {
		return errors.New("password must not contain a NUL byte")
	}
	return nil
}
