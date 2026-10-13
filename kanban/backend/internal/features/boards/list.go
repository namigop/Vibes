package boards

import (
	"net/http"

	"kanban/internal/domain"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// listBoards returns the caller's boards in sidebar order.
//
// The owner filter lives in the query, not here: there is deliberately no code
// path that lists boards without also constraining them to one user.
func listBoards(q *store.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := caller(w, r)
		if !ok {
			return
		}

		rows, err := q.ListBoards(r.Context(), owner)
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		httpserver.WriteJSON(w, http.StatusOK, domain.BoardsFromStore(rows))
	}
}
