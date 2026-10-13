package database

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

// IsNotFound is the only thing standing between a sqlc :one query that matched
// nothing and a 500. Every slice that fetches by id funnels through it, so its
// handling of the non-ErrNoRows cases matters as much as the positive one.
func TestIsNotFound(t *testing.T) {
	t.Run("pgx.ErrNoRows is not found", func(t *testing.T) {
		if !IsNotFound(pgx.ErrNoRows) {
			t.Error("ErrNoRows should be reported as not found")
		}
	})

	t.Run("a wrapped ErrNoRows is not found", func(t *testing.T) {
		// A slice that adds context with %w keeps the sentinel reachable, so
		// errors.Is must succeed. database.IsNotFound uses ==, so this asserts
		// the boundary: unwrapping is the slice's job, not this predicate's.
		wrapped := errors.New("get board: " + pgx.ErrNoRows.Error())
		if IsNotFound(wrapped) {
			t.Error("a string-wrapped error is not the sentinel and must not match")
		}
	})

	t.Run("other errors are not not-found", func(t *testing.T) {
		if IsNotFound(errors.New("connection reset")) {
			t.Error("a transport error must not be mistaken for a missing row")
		}
	})

	t.Run("nil is not not-found", func(t *testing.T) {
		if IsNotFound(nil) {
			t.Error("nil must not be reported as not found")
		}
	})
}
