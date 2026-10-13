package cards

import (
	"net/http"

	"kanban/internal/domain"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// updateCard replaces a card's editable fields.
//
// The column and the ordering are not editable here: moving is its own
// operation with its own invariants, and folding it in would let a plain edit
// silently reorder a column.
func updateCard(q *store.Queries) http.HandlerFunc {
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

		var req CardRequest
		if err := httpserver.DecodeJSON(r, &req); err != nil {
			httpserver.WriteError(w, err)
			return
		}

		f, err := validate(req)
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		due, err := domain.DueDateIn(f.dueDate)
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		card, err := q.UpdateCard(r.Context(), store.UpdateCardParams{
			ID:          id,
			Title:       f.title,
			Description: f.description,
			Assignee:    f.assignee,
			Priority:    f.priority,
			DueDate:     due,
			OwnerID:     owner,
		})
		if err != nil {
			httpserver.WriteError(w, notFoundIfMissing(err, "card"))
			return
		}

		httpserver.WriteJSON(w, http.StatusOK, domain.CardFromStore(card))
	}
}
