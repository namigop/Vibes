package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/onezerotoys/todo-app/backend/internal/todo"
)

// TodosHandler holds dependencies for the /todos endpoints.
type TodosHandler struct {
	svc   *todo.Service
	valid *todo.Validator
	log   *slog.Logger
}

// NewTodosHandler wires the handler with its collaborators.
func NewTodosHandler(svc *todo.Service, valid *todo.Validator, log *slog.Logger) *TodosHandler {
	return &TodosHandler{svc: svc, valid: valid, log: log}
}

// List handles GET /todos.
func (h *TodosHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		h.log.Error("list todos failed", slog.Any("err", err))
		WriteInternal(w)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

// Get handles GET /todos/{id}.
func (h *TodosHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	t, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, t)
}

// Create handles POST /todos.
func (h *TodosHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req todo.CreateRequest
	if !DecodeJSON(w, r, &req) {
		return
	}
	if details := h.valid.Validate(&req); details != nil {
		WriteError(w, http.StatusBadRequest, "validation_failed", "request body is invalid", details...)
		return
	}
	t, err := h.svc.Create(r.Context(), req)
	if err != nil {
		h.log.Error("create todo failed", slog.Any("err", err))
		WriteInternal(w)
		return
	}
	WriteJSON(w, http.StatusCreated, t)
}

// Update handles PATCH /todos/{id}.
func (h *TodosHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req todo.UpdateRequest
	if !DecodeJSON(w, r, &req) {
		return
	}
	if details := h.valid.Validate(&req); details != nil {
		WriteError(w, http.StatusBadRequest, "validation_failed", "request body is invalid", details...)
		return
	}
	t, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, t)
}

// Delete handles DELETE /todos/{id}.
func (h *TodosHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TodosHandler) handleServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, todo.ErrNotFound) {
		WriteError(w, http.StatusNotFound, "not_found", "todo not found")
		return
	}
	h.log.Error("todo service error", slog.Any("err", err))
	WriteInternal(w)
}

// parseID extracts a UUID from the {id} chi URL param. On failure it writes
// a 400 envelope and returns ok=false.
func parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	raw := chi.URLParam(r, "id")
	id, err := uuid.Parse(raw)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "id must be a valid UUID")
		return uuid.Nil, false
	}
	return id, true
}
