package todo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/onezerotoys/todo-app/backend/internal/store"
)

// Store is the persistence interface used by the service.
// We keep it small so handlers can be tested with fakes; the sqlc-generated
// *store.Queries satisfies this contract.
type Store interface {
	ListTodos(ctx context.Context) ([]store.Todo, error)
	GetTodo(ctx context.Context, id uuid.UUID) (store.Todo, error)
	CreateTodo(ctx context.Context, arg store.CreateTodoParams) (store.Todo, error)
	UpdateTodo(ctx context.Context, arg store.UpdateTodoParams) (store.Todo, error)
	DeleteTodo(ctx context.Context, id uuid.UUID) (int64, error)
}

// Service is the business logic boundary between handlers and storage.
type Service struct {
	store Store
}

// NewService wires a Service around the given store.
func NewService(s Store) *Service {
	return &Service{store: s}
}

// ErrNotFound is returned when the requested todo does not exist.
var ErrNotFound = errors.New("todo not found")

// List returns all todos ordered by created_at desc.
// The slice is normalised to non-nil so the API serialises an empty list as
// [] rather than null, which is what the frontend expects.
func (s *Service) List(ctx context.Context) ([]store.Todo, error) {
	items, err := s.store.ListTodos(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []store.Todo{}
	}
	return items, nil
}

// Get returns a single todo by id or ErrNotFound.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (store.Todo, error) {
	t, err := s.store.GetTodo(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.Todo{}, ErrNotFound
		}
		return store.Todo{}, err
	}
	return t, nil
}

// Create persists a new todo.
func (s *Service) Create(ctx context.Context, req CreateRequest) (store.Todo, error) {
	return s.store.CreateTodo(ctx, store.CreateTodoParams{
		Title:       req.Title,
		Description: req.Description,
	})
}

// Update applies a partial update. Returns ErrNotFound if the row is gone.
func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (store.Todo, error) {
	t, err := s.store.UpdateTodo(ctx, store.UpdateTodoParams{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Done:        req.Done,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.Todo{}, ErrNotFound
		}
		return store.Todo{}, err
	}
	return t, nil
}

// Delete removes a todo. Returns ErrNotFound if the row was already gone.
//
// The query is :execrows, so it reports how many rows were affected rather
// than returning pgx.ErrNoRows. Zero affected rows means the id did not
// exist, which we surface as ErrNotFound so the handler returns 404.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	affected, err := s.store.DeleteTodo(ctx, id)
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
