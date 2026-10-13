package columns

import (
	"net/http"

	"kanban/internal/domain"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// deleteColumn removes a column and every card inside it.
//
// The delete reports rows-affected rather than using `:exec`, because a no-op
// DELETE does not error — without the count, deleting somebody else's column
// would silently return 204.
func deleteColumn(q *store.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := caller(w, r)
		if !ok {
			return
		}

		id, err := httpserver.PathUUID(r, "columnId", "column id")
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		affected, err := q.DeleteColumn(r.Context(), store.DeleteColumnParams{
			ID: id, OwnerID: owner,
		})
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}
		if affected == 0 {
			httpserver.WriteError(w, domain.NotFound("column not found"))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
