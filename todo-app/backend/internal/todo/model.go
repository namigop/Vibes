package todo

import (
	"time"

	"github.com/google/uuid"
)

// Todo is the domain representation of a todo item.
type Todo struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateRequest is the payload for POST /todos.
type CreateRequest struct {
	Title       string `json:"title" validate:"required,max=200"`
	Description string `json:"description" validate:"omitempty,max=2000"`
}

// UpdateRequest is the payload for PATCH /todos/{id}.
// All fields are optional; nil pointer means "do not change".
type UpdateRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,max=200"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	Done        *bool   `json:"done,omitempty"`
}
