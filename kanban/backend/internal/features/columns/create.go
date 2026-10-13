package columns

import (
	"net/http"

	"kanban/internal/domain"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// CreateColumnRequest is the body for POST /boards/{id}/columns.
type CreateColumnRequest struct {
	Name string `json:"name"`
}

// createColumn appends a custom deal stage to the end of a board.
func createColumn(q *store.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := caller(w, r)
		if !ok {
			return
		}

		boardID, err := httpserver.PathUUID(r, "boardId", "board id")
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		var req CreateColumnRequest
		if err := httpserver.DecodeJSON(r, &req); err != nil {
			httpserver.WriteError(w, err)
			return
		}

		name, err := domain.RequireName(req.Name, "name")
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		// Check ownership before inserting, so creating a column on somebody
		// else's board is a 404 rather than a silent success.
		exists, err := q.BoardExists(r.Context(), store.BoardExistsParams{
			ID: boardID, OwnerID: owner,
		})
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}
		if !exists {
			httpserver.WriteError(w, domain.NotFound("board not found"))
			return
		}

		col, err := q.CreateColumn(r.Context(), store.CreateColumnParams{
			BoardID: boardID,
			Name:    name,
		})
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		httpserver.WriteJSON(w, http.StatusCreated, domain.ColumnFromStore(col))
	}
}
