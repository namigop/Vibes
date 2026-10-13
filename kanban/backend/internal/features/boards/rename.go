package boards

import (
	"net/http"

	"kanban/internal/domain"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// RenameBoardRequest is the body for PATCH /boards/{id}.
type RenameBoardRequest struct {
	Name string `json:"name"`
}

// renameBoard changes a board's name.
func renameBoard(q *store.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := caller(w, r)
		if !ok {
			return
		}

		id, err := httpserver.PathUUID(r, "boardId", "board id")
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		var req RenameBoardRequest
		if err := httpserver.DecodeJSON(r, &req); err != nil {
			httpserver.WriteError(w, err)
			return
		}

		name, err := domain.RequireName(req.Name, "name")
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		// OwnerID is part of the WHERE clause, so renaming somebody else's board
		// updates nothing and comes back as a 404.
		board, err := q.UpdateBoardName(r.Context(), store.UpdateBoardNameParams{
			ID:      id,
			Name:    name,
			OwnerID: owner,
		})
		if err != nil {
			httpserver.WriteError(w, notFoundIfMissing(err, "board"))
			return
		}

		httpserver.WriteJSON(w, http.StatusOK, domain.BoardFromStore(board))
	}
}
