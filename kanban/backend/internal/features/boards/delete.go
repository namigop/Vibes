package boards

import (
	"net/http"

	"kanban/internal/domain"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// deleteBoard removes a board; its columns and cards cascade in the database.
//
// The delete reports rows-affected rather than using `:exec`, because a
// no-op DELETE does not error — without the count, deleting somebody else's
// board would silently return 204 and the client would drop a board that still
// exists.
func deleteBoard(q *store.Queries) http.HandlerFunc {
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

		affected, err := q.DeleteBoard(r.Context(), store.DeleteBoardParams{
			ID:      id,
			OwnerID: owner,
		})
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}
		if affected == 0 {
			httpserver.WriteError(w, domain.NotFound("board not found"))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
