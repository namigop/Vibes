// Package columns contains the column slices: create, rename, move, delete.
//
// A column is a workflow stage — "To Do", "Negotiation", whatever the board
// needs. Reordering is the interesting operation: see move.go.
package columns

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

// Register mounts every column route.
//
// The pool is needed alongside the queries because reordering runs in a
// transaction, and store.Queries does not carry one. Slices that only read can
// ignore it.
func Register(r *httpserver.Router, q *store.Queries, pool *pgxpool.Pool) {
	r.Private("POST /api/v1/boards/{boardId}/columns", createColumn(q))
	r.Private("PATCH /api/v1/columns/{columnId}", renameColumn(q))
	r.Private("POST /api/v1/columns/{columnId}/move", moveColumn(q, pool))
	r.Private("DELETE /api/v1/columns/{columnId}", deleteColumn(q))
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

// writePositions reindexes one column to a gap-free ordering.
//
// reassign is true when the item is moving into this column, so it must be
// parented first; otherwise the positions it is about to receive would belong
// to a different parent.
func writePositions(
	ctx context.Context,
	q *store.Queries,
	ids []uuid.UUID,
	itemID uuid.UUID,
	columnID uuid.UUID,
	owner uuid.UUID,
	reassign bool,
) error {
	if reassign {
		err := q.SetCardColumn(ctx, store.SetCardColumnParams{
			ID: itemID, ColumnID: columnID, OwnerID: owner,
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
