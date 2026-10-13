package domain

// Error kinds live here rather than in the HTTP layer so that a slice can fail
// meaningfully without importing anything web-shaped. platform/httpserver
// translates them into status codes; domain stays ignorant of HTTP.
//
// A sentinel-ish approach is deliberately not used: each kind carries a
// human-readable message because that message is shown in a toast and must
// say which field was wrong.

import "fmt"

// ValidationError is a client input problem -> HTTP 400.
type ValidationError struct{ Message string }

func (e ValidationError) Error() string { return e.Message }

// NotFoundError means the addressed resource does not exist, or belongs to
// someone else -> HTTP 404.
//
// A slice uses this for both cases on purpose. Answering 403 to "that board is
// not yours" would confirm the board exists, leaking the fact that another user
// has one.
type NotFoundError struct{ Message string }

func (e NotFoundError) Error() string { return e.Message }

// ConflictError -> HTTP 409.
type ConflictError struct{ Message string }

func (e ConflictError) Error() string { return e.Message }

// UnauthorizedError -> HTTP 401.
type UnauthorizedError struct{ Message string }

func (e UnauthorizedError) Error() string { return e.Message }

// Invalid builds a ValidationError with a formatted message.
func Invalid(format string, args ...any) error {
	return ValidationError{Message: fmt.Sprintf(format, args...)}
}

// NotFound builds a NotFoundError with a formatted message.
func NotFound(format string, args ...any) error {
	return NotFoundError{Message: fmt.Sprintf(format, args...)}
}

// Conflict builds a ConflictError with a formatted message.
func Conflict(format string, args ...any) error {
	return ConflictError{Message: fmt.Sprintf(format, args...)}
}

// Unauthorized builds an UnauthorizedError with a formatted message.
func Unauthorized(format string, args ...any) error {
	return UnauthorizedError{Message: fmt.Sprintf(format, args...)}
}
