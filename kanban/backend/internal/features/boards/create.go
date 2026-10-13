package boards

import (
	"net/http"

	"kanban/internal/domain"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// CreateBoardRequest is the body for POST /boards.
type CreateBoardRequest struct {
	Name string `json:"name"`
}

// createBoard adds a board owned by the caller, seeded with the default stages.
//
// The stages are created here rather than by a database default or a trigger so
// that "what a new board looks like" is answerable by reading one function.
func createBoard(q *store.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := caller(w, r)
		if !ok {
			return
		}

		var req CreateBoardRequest
		if err := httpserver.DecodeJSON(r, &req); err != nil {
			httpserver.WriteError(w, err)
			return
		}

		name, err := domain.RequireName(req.Name, "name")
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		board, err := q.CreateBoard(r.Context(), store.CreateBoardParams{
			Name:    name,
			OwnerID: owner,
		})
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		for _, stage := range domain.DefaultColumnNames {
			_, err := q.CreateColumn(r.Context(), store.CreateColumnParams{
				BoardID: board.ID,
				Name:    stage,
			})
			if err != nil {
				httpserver.WriteError(w, err)
				return
			}
		}

		httpserver.WriteJSON(w, http.StatusCreated, domain.BoardFromStore(board))
	}
}
