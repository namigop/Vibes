package cards

import (
	"net/http"

	"kanban/internal/domain"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// createCard appends a card to the end of a column.
func createCard(q *store.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := caller(w, r)
		if !ok {
			return
		}

		columnID, err := httpserver.PathUUID(r, "columnId", "column id")
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

		// A card must hang off a column the caller owns. Checked before the
		// insert so creating a card in somebody else's column is a 404.
		if _, err := q.GetColumn(r.Context(), store.GetColumnParams{
			ID: columnID, OwnerID: owner,
		}); err != nil {
			httpserver.WriteError(w, notFoundIfMissing(err, "column"))
			return
		}

		due, err := domain.DueDateIn(f.dueDate)
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		card, err := q.CreateCard(r.Context(), store.CreateCardParams{
			ColumnID:    columnID,
			Title:       f.title,
			Description: f.description,
			Assignee:    f.assignee,
			Priority:    f.priority,
			DueDate:     due,
		})
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		httpserver.WriteJSON(w, http.StatusCreated, domain.CardFromStore(card))
	}
}
