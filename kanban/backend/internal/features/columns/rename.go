package columns

import (
	"net/http"

	"kanban/internal/domain"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// RenameColumnRequest is the body for PATCH /columns/{id}.
type RenameColumnRequest struct {
	Name string `json:"name"`
}

// renameColumn changes a column's name.
func renameColumn(q *store.Queries) http.HandlerFunc {
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

		var req RenameColumnRequest
		if err := httpserver.DecodeJSON(r, &req); err != nil {
			httpserver.WriteError(w, err)
			return
		}

		name, err := domain.RequireName(req.Name, "name")
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		col, err := q.UpdateColumnName(r.Context(), store.UpdateColumnNameParams{
			ID: id, Name: name, OwnerID: owner,
		})
		if err != nil {
			httpserver.WriteError(w, notFoundIfMissing(err, "column"))
			return
		}

		httpserver.WriteJSON(w, http.StatusOK, domain.ColumnFromStore(col))
	}
}
