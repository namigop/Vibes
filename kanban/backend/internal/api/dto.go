// Package api holds the HTTP layer: request/response DTOs, the router, the
// CORS middleware and the request handlers.
//
// The DTOs in this file are the Go side of the frozen contract in
// docs/API.md. They are deliberately separate from the sqlc-generated models
// so that the wire format stays stable even if the database schema changes.
package api

import "time"

// ErrorResponse is the single error envelope every non-2xx response uses.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Stable machine-readable error slugs.
const (
	ErrValidation   = "validation_failed"
	ErrNotFound     = "not_found"
	ErrConflict     = "conflict"
	ErrInternal     = "internal_error"
	ErrUnauthorized = "unauthorized"
)

// User is the public view of an account. It never includes password_hash — the
// DTO is what gets serialised, so the hash cannot leak by omission.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// Board is a collection of columns owned by exactly one user.
type Board struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Position  int32     `json:"position"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Column is one workflow stage within a board (e.g. "In Progress").
type Column struct {
	ID        string    `json:"id"`
	BoardID   string    `json:"board_id"`
	Name      string    `json:"name"`
	Position  int32     `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Card is one unit of work. DueDate is a calendar date rendered as
// "2006-01-02" in UTC, or nil when the card has no due date.
//
// Assignee is free text naming a person involved in the work, NOT an account —
// boards are owner-only, so a user reference here would name someone who
// cannot open the card.
type Card struct {
	ID          string    `json:"id"`
	ColumnID    string    `json:"column_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Assignee    string    `json:"assignee"`
	Priority    string    `json:"priority"`
	DueDate     *string   `json:"due_date"`
	Position    int32     `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BoardDetail is the single response that hydrates a whole board.
type BoardDetail struct {
	Board   Board    `json:"board"`
	Columns []Column `json:"columns"`
	Cards   []Card   `json:"cards"`
}

// HealthResponse reports API and database liveness.
type HealthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

// ---- request bodies -------------------------------------------------------

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse carries the bearer token and the signed-in user.
type LoginResponse struct {
	Token     string    `json:"token"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}

// CreateBoardRequest is the body for POST /boards.
type CreateBoardRequest struct {
	Name string `json:"name"`
}

// RenameBoardRequest is the body for PATCH /boards/{id}.
type RenameBoardRequest struct {
	Name string `json:"name"`
}

// CreateColumnRequest is the body for POST /boards/{boardId}/columns.
type CreateColumnRequest struct {
	Name string `json:"name"`
}

// RenameColumnRequest is the body for PATCH /columns/{id}.
type RenameColumnRequest struct {
	Name string `json:"name"`
}

// MoveColumnRequest is the body for POST /columns/{id}/move. Position is the
// destination index within the board's column list.
type MoveColumnRequest struct {
	Position int32 `json:"position"`
}

// CardRequest is the body for both POST /columns/{id}/cards and
// PATCH /cards/{id} — the API treats card updates as a full replace.
//
// DueDate is a *string so that an explicit JSON null can be distinguished
// from an omitted field; both mean "no due date".
type CardRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Assignee    *string `json:"assignee"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"due_date"`
}

// MoveCardRequest is the body for POST /cards/{id}/move. Position is the
// destination index within the destination column *after* the move.
type MoveCardRequest struct {
	ColumnID string `json:"column_id"`
	Position int32  `json:"position"`
}
