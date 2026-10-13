package store

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// isNoRows reports whether err is pgx's "no rows in result set".
func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// hashPassword wraps bcrypt so seed.go does not import it directly.
func hashPassword(plain string) (string, error) {
	digest, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(digest), nil
}

// dueOffset builds a calendar date `n` days from today. A zero offset means
// "no due date", which the demo data uses for cards that are not time-bound.
func dueOffset(today time.Time, n int) pgtype.Date {
	if n == 0 {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: today.AddDate(0, 0, n), Valid: true}
}
