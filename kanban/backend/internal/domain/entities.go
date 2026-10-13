// Package domain holds the entities, invariants and error kinds that more than
// one feature slice shares.
//
// It has no knowledge of HTTP and no knowledge of which slice is calling it.
// A feature depends on domain and store; domain depends on neither.
package domain

import "time"

// User is the public view of an account. It never carries password_hash, so
// the hash cannot leak by omission.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// Board belongs to exactly one user. OwnerID is non-empty on every board
// because the schema enforces NOT NULL.
type Board struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Position  int32     `json:"position"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Column is one workflow stage within a board.
type Column struct {
	ID        string    `json:"id"`
	BoardID   string    `json:"board_id"`
	Name      string    `json:"name"`
	Position  int32     `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Card is one unit of work.
//
// DueDate is a calendar date rendered as "2006-01-02", or nil.
//
// Assignee is free text and deliberately NOT a user reference. Boards are
// owner-only, so the only person who can open a card is its owner; a user
// reference would name someone who by construction cannot see the card. It
// names a person involved in the work instead — a deal contact, a colleague
// with no account.
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

// BoardDetail is the single response that hydrates a whole board screen.
type BoardDetail struct {
	Board   Board    `json:"board"`
	Columns []Column `json:"columns"`
	Cards   []Card   `json:"cards"`
}

// Health reports API and database liveness.
type Health struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}
