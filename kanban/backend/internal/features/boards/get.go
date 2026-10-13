package boards

import (
	"net/http"

	"kanban/internal/domain"
	"kanban/internal/platform/httpserver"
	"kanban/internal/store"
)

// getBoard hydrates a whole board screen: the board, its columns and its cards.
//
// One request instead of three, because the frontend renders all three at once
// and a partially-loaded board is worse than a slightly slower one.
//
// A board belonging to somebody else is reported as not_found rather than
// forbidden — a 403 would confirm the board exists.
func getBoard(q *store.Queries) http.HandlerFunc {
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

		board, err := q.GetBoard(r.Context(), store.GetBoardParams{ID: id, OwnerID: owner})
		if err != nil {
			httpserver.WriteError(w, notFoundIfMissing(err, "board"))
			return
		}

		columns, err := q.ListColumns(r.Context(),
			store.ListColumnsParams{BoardID: id, OwnerID: owner})
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		cards, err := q.ListCardsForBoard(r.Context(),
			store.ListCardsForBoardParams{BoardID: id, OwnerID: owner})
		if err != nil {
			httpserver.WriteError(w, err)
			return
		}

		httpserver.WriteJSON(w, http.StatusOK, domain.BoardDetail{
			Board:   domain.BoardFromStore(board),
			Columns: domain.ColumnsFromStore(columns),
			Cards:   domain.CardsFromStore(cards),
		})
	}
}
