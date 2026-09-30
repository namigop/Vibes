package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/onezerotoys/todo-app/backend/internal/todo"
)

// NewRouter wires the chi mux with middleware and mounts all routes.
func NewRouter(svc *todo.Service, valid *todo.Validator, logger *slog.Logger, corsOrigins []string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(RequestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(CORS(corsOrigins))

	r.Get("/healthz", HealthHandler)

	r.Route("/api/v1", func(r chi.Router) {
		h := NewTodosHandler(svc, valid, logger)
		r.Get("/todos", h.List)
		r.Post("/todos", h.Create)
		r.Get("/todos/{id}", h.Get)
		r.Patch("/todos/{id}", h.Update)
		r.Delete("/todos/{id}", h.Delete)
	})

	return r
}
