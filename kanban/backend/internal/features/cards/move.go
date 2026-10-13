package cards

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"kanban/internal/domain"
	"kanban/internal/platform/database"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// MoveCardRequest is the body for POST /cards/{id}/move.
//
// Position is the index within the destination column *after* the move, and it
// is clamped rather than rejected so a stale drag still lands sensibly.
type MoveCardRequest struct {
	ColumnID string `json:"column_id"`
	Position int32  `json:"position"`
}

// moveCard moves a card within a column or across columns, and persists it.
//
// This is the endpoint behind every drag. Two things make it non-trivial:
//
//   - Both the source and the destination column are reindexed, so a card
//     leaving column A leaves no gap behind it. Reindexing in place would
//     instead need fractional positions, which drift over time.
//   - It all happens in one transaction. A partial reindex would leave a column
//     with duplicate or missing positions, which is worse than the drag failing.
//
// The response is the settled card, carrying the position the server actually
// chose — which may differ from the requested one, because the index is
// clamped against the real column length.
func moveCard(q *store.Queries, pool *pgxpool.Pool) http.HandlerFunc {
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

		var req MoveCardRequest
		if err := httpserver.DecodeJSON(r, &req); err != nil {
			httpserver.WriteError(w, err)
			return
		}

		columnID, err := httpserver.ParseUUID(req.ColumnID, "column_id")
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		var moved domain.Card

		err = database.UnitOfWork(r.Context(), pool, func(tq *store.Queries) error {
			card, err := tq.GetCard(r.Context(), store.GetCardParams{ID: id, OwnerID: owner})
			if err != nil {
				return notFoundIfMissing(err, "card")
			}

			if _, err := tq.GetColumn(r.Context(), store.GetColumnParams{
				ID: columnID, OwnerID: owner,
			}); err != nil {
				return notFoundIfMissing(err, "column")
			}

			if card.ColumnID == columnID {
				// Reorder in place: the card is already in the destination list,
				// so removing it first avoids an off-by-one in the clamp.
				ids, err := tq.ListCardIDsInColumn(r.Context(),
					store.ListCardIDsInColumnParams{ColumnID: columnID, OwnerID: owner})
				if err != nil {
					return err
				}

				next := domain.InsertAt(
					domain.RemoveID(ids, id),
					domain.ClampIndex(req.Position, int32(len(ids)-1)),
					id,
				)
				if err := writePositions(r.Context(), tq, next, id, columnID, owner, false); err != nil {
					return err
				}
			} else {
				sourceIDs, err := tq.ListCardIDsInColumn(r.Context(),
					store.ListCardIDsInColumnParams{ColumnID: card.ColumnID, OwnerID: owner})
				if err != nil {
					return err
				}
				destIDs, err := tq.ListCardIDsInColumn(r.Context(),
					store.ListCardIDsInColumnParams{ColumnID: columnID, OwnerID: owner})
				if err != nil {
					return err
				}

				nextSource := domain.RemoveID(sourceIDs, id)
				nextDest := domain.InsertAt(
					destIDs, domain.ClampIndex(req.Position, int32(len(destIDs))), id)

				if err := writePositions(r.Context(), tq, nextSource, id, card.ColumnID, owner, false); err != nil {
					return err
				}
				if err := writePositions(r.Context(), tq, nextDest, id, columnID, owner, true); err != nil {
					return err
				}
			}

			settled, err := tq.GetCard(r.Context(), store.GetCardParams{ID: id, OwnerID: owner})
			if err != nil {
				return err
			}
			moved = domain.CardFromStore(settled)
			return nil
		})
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		httpserver.WriteJSON(w, http.StatusOK, moved)
	}
}
