// Package cards contains the card slices: create, update, move, delete.
//
// Cards are the unit of drag-and-drop, so move.go is the substantial one here.
package cards

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"kanban/internal/domain"
	"kanban/internal/platform/auth"
	"kanban/internal/platform/database"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// Register mounts every card route.
//
// The pool is needed alongside the queries because move reindexes across two
// columns inside a transaction, and store.Queries does not carry one.
func Register(r *httpserver.Router, q *store.Queries, pool *pgxpool.Pool) {
	r.Private("POST /api/v1/columns/{columnId}/cards", createCard(q))
	r.Private("PATCH /api/v1/cards/{cardId}", updateCard(q))
	r.Private("POST /api/v1/cards/{cardId}/move", moveCard(q, pool))
	r.Private("DELETE /api/v1/cards/{cardId}", deleteCard(q))
}

// caller resolves the authenticated user, writing a 401 and reporting false if
// there is none.
func caller(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, ok := auth.CallerFrom(r.Context())
	if !ok {
		httpserver.WriteError(w, auth.ErrInvalidToken)
		return uuid.Nil, false
	}
	return id, true
}

// notFoundIfMissing turns sqlc's ErrNoRows into a domain 404 and leaves every
// other error untouched.
func notFoundIfMissing(err error, what string) error {
	if database.IsNotFound(err) {
		return domain.NotFound("%s not found", what)
	}
	return err
}

// CardRequest is the body for both create and full update.
//
// The API treats a card update as a full replace, so every field is present in
// the type even though only title is required. DueDate is a pointer so that an
// explicit JSON null can be told apart from an omitted field; both mean "no due
// date".
type CardRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Assignee    *string `json:"assignee"`
	Priority    string  `json:"priority"`
	DueDate     *string `json:"due_date"`
}

// fields is a validated, normalised CardRequest.
type fields struct {
	title       string
	description *string
	assignee    *string
	priority    string
	dueDate     *string
}

// validate normalises and checks a card body.
func validate(req CardRequest) (fields, error) {
	title, err := domain.RequireName(req.Title, "title")
	if err != nil {
		return fields{}, err
	}

	priority, err := domain.RequirePriority(req.Priority)
	if err != nil {
		return fields{}, err
	}

	return fields{
		title:       title,
		description: domain.NormaliseText(req.Description),
		assignee:    domain.NormaliseText(req.Assignee),
		priority:    priority,
		dueDate:     req.DueDate,
	}, nil
}

// writePositions reindexes one column to a gap-free ordering.
//
// reassign is true when the card is moving into this column, so it must be
// parented first; otherwise the positions it is about to receive would belong
// to a different parent.
func writePositions(
	ctx context.Context,
	q *store.Queries,
	ids []uuid.UUID,
	cardID uuid.UUID,
	columnID uuid.UUID,
	owner uuid.UUID,
	reassign bool,
) error {
	if reassign {
		err := q.SetCardColumn(ctx, store.SetCardColumnParams{
			ID: cardID, ColumnID: columnID, OwnerID: owner,
		})
		if err != nil {
			return err
		}
	}

	for i, id := range ids {
		err := q.SetCardPosition(ctx, store.SetCardPositionParams{
			ID: id, Position: int32(i), OwnerID: owner,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
