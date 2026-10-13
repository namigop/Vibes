package columns

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"kanban/internal/domain"
	"kanban/internal/platform/database"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// MoveColumnRequest is the body for POST /columns/{id}/move.
//
// Position is the destination index within the board's column list, and it is
// clamped rather than rejected so a slightly stale drag still lands sensibly.
type MoveColumnRequest struct {
	Position int32 `json:"position"`
}

// moveColumn reorders a column within its board.
//
// Every column of the board is reindexed inside one transaction, and the full
// reordered list comes back so the client can replace its local state with the
// authoritative one instead of guessing.
func moveColumn(q *store.Queries, pool *pgxpool.Pool) http.HandlerFunc {
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

		var req MoveColumnRequest
		if err := httpserver.DecodeJSON(r, &req); err != nil {
			httpserver.WriteError(w, err)
			return
		}

		var ordered []domain.Column

		err = database.UnitOfWork(r.Context(), pool, func(tq *store.Queries) error {
			col, err := tq.GetColumn(r.Context(), store.GetColumnParams{ID: id, OwnerID: owner})
			if err != nil {
				return notFoundIfMissing(err, "column")
			}

			ids, err := tq.ListColumnIDs(r.Context(), store.ListColumnIDsParams{
				BoardID: col.BoardID, OwnerID: owner,
			})
			if err != nil {
				return err
			}

			next := domain.InsertAt(
				domain.RemoveID(ids, id),
				domain.ClampIndex(req.Position, int32(len(ids)-1)),
				id,
			)

			for i, cid := range next {
				err := tq.SetColumnOrdinal(r.Context(), store.SetColumnOrdinalParams{
					ID: cid, Position: int32(i), OwnerID: owner,
				})
				if err != nil {
					return err
				}
			}

			cols, err := tq.ListColumns(r.Context(), store.ListColumnsParams{
				BoardID: col.BoardID, OwnerID: owner,
			})
			if err != nil {
				return err
			}
			ordered = domain.ColumnsFromStore(cols)
			return nil
		})
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		httpserver.WriteJSON(w, http.StatusOK, ordered)
	}
}
