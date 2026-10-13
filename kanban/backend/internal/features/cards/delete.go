package cards

import (
	"net/http"

	"kanban/internal/domain"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// deleteCard removes a single card.
//
// The delete reports rows-affected rather than using `:exec`, because a no-op
// DELETE does not error — without the count, deleting somebody else's card
// would silently return 204.
func deleteCard(q *store.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := caller(w, r)
		if !ok {
			return
		}

		id, err := httpserver.PathUUID(r, "cardId", "card id")
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		affected, err := q.DeleteCard(r.Context(), store.DeleteCardParams{
			ID: id, OwnerID: owner,
		})
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}
		if affected == 0 {
			httpserver.WriteError(w, domain.NotFound("card not found"))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
