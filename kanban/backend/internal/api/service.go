package api

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"kanban/internal/store"
)

// Service holds the business rules that sit between HTTP and SQL.
//
// The interesting one is ordering: `position` is a gap-free ordinal, and any
// move reindexes both the source and the destination in a single transaction.
// That makes ordering impossible to corrupt with a stale client, and avoids
// the drift that fractional "insert between 3 and 7" schemes accumulate.
type Service struct {
	q    *store.Queries
	pool *pgxpool.Pool
}

// NewService wires a service to its store and connection pool.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{q: store.New(pool), pool: pool}
}

// DefaultColumnNames are the workflow stages every new board starts with.
var DefaultColumnNames = []string{"To Do", "In Progress", "Review", "Done"}

// validPriorities is the closed set the database CHECK constraint enforces.
// It is repeated here so the API can answer 400 with a useful message instead
// of leaking a Postgres constraint violation as a 500.
var validPriorities = map[string]struct{}{
	"low": {}, "medium": {}, "high": {}, "urgent": {},
}

const defaultPriority = "medium"

// ---- boards ---------------------------------------------------------------

// ListBoards returns every board, in sidebar order.
func (s *Service) ListBoards(ctx context.Context) ([]Board, error) {
	boards, err := s.q.ListBoards(ctx)
	if err != nil {
		return nil, err
	}
	return boardDTOs(boards), nil
}

// GetBoardDetail loads a whole board — the board, its columns and every card —
// in the single request the frontend uses to hydrate a board screen.
func (s *Service) GetBoardDetail(ctx context.Context, id uuid.UUID) (BoardDetail, error) {
	board, err := s.q.GetBoard(ctx, id)
	if err != nil {
		return BoardDetail{}, mapNotFound(err, "board")
	}

	columns, err := s.q.ListColumns(ctx, id)
	if err != nil {
		return BoardDetail{}, err
	}
	cards, err := s.q.ListCardsForBoard(ctx, id)
	if err != nil {
		return BoardDetail{}, err
	}

	return BoardDetail{
		Board:   boardDTO(board),
		Columns: columnDTOs(columns),
		Cards:   cardDTOs(cards),
	}, nil
}

// CreateBoard adds a board and seeds the four default workflow stages.
func (s *Service) CreateBoard(ctx context.Context, rawName string) (Board, error) {
	name, err := requireName(rawName, "name")
	if err != nil {
		return Board{}, err
	}

	board, err := s.q.CreateBoard(ctx, name)
	if err != nil {
		return Board{}, err
	}

	for _, stage := range DefaultColumnNames {
		_, err := s.q.CreateColumn(ctx, store.CreateColumnParams{
			BoardID: board.ID,
			Name:    stage,
		})
		if err != nil {
			return Board{}, err
		}
	}

	return boardDTO(board), nil
}

// RenameBoard changes a board's name.
func (s *Service) RenameBoard(ctx context.Context, id uuid.UUID, rawName string) (Board, error) {
	name, err := requireName(rawName, "name")
	if err != nil {
		return Board{}, err
	}

	board, err := s.q.UpdateBoardName(ctx, store.UpdateBoardNameParams{ID: id, Name: name})
	if err != nil {
		return Board{}, mapNotFound(err, "board")
	}
	return boardDTO(board), nil
}

// DeleteBoard removes a board; its columns and cards cascade in the database.
func (s *Service) DeleteBoard(ctx context.Context, id uuid.UUID) error {
	affected, err := s.q.DeleteBoard(ctx, id)
	if err != nil {
		return err
	}
	if affected == 0 {
		return notFound("board not found")
	}
	return nil
}

// ---- columns --------------------------------------------------------------

// CreateColumn appends a custom column (a deal stage) to a board.
func (s *Service) CreateColumn(ctx context.Context, boardID uuid.UUID, rawName string) (Column, error) {
	name, err := requireName(rawName, "name")
	if err != nil {
		return Column{}, err
	}

	exists, err := s.q.BoardExists(ctx, boardID)
	if err != nil {
		return Column{}, err
	}
	if !exists {
		return Column{}, notFound("board not found")
	}

	col, err := s.q.CreateColumn(ctx, store.CreateColumnParams{BoardID: boardID, Name: name})
	if err != nil {
		return Column{}, err
	}
	return columnDTO(col), nil
}

// RenameColumn changes a column's name.
func (s *Service) RenameColumn(ctx context.Context, id uuid.UUID, rawName string) (Column, error) {
	name, err := requireName(rawName, "name")
	if err != nil {
		return Column{}, err
	}

	col, err := s.q.UpdateColumnName(ctx, store.UpdateColumnNameParams{ID: id, Name: name})
	if err != nil {
		return Column{}, mapNotFound(err, "column")
	}
	return columnDTO(col), nil
}

// MoveColumn reorders a column within its board and returns every column of
// that board in the new order, so the client can replace its local state with
// the authoritative one.
func (s *Service) MoveColumn(ctx context.Context, id uuid.UUID, position int32) ([]Column, error) {
	var ordered []Column

	err := s.withTx(ctx, func(q *store.Queries) error {
		col, err := q.GetColumn(ctx, id)
		if err != nil {
			return mapNotFound(err, "column")
		}

		ids, err := q.ListColumnIDs(ctx, col.BoardID)
		if err != nil {
			return err
		}

		next := insertAt(removeID(ids, id), clampIndex(position, int32(len(ids)-1)), id)

		for i, cid := range next {
			err := q.SetColumnOrdinal(ctx, store.SetColumnOrdinalParams{
				ID:       cid,
				Position: int32(i),
			})
			if err != nil {
				return err
			}
		}

		cols, err := q.ListColumns(ctx, col.BoardID)
		if err != nil {
			return err
		}
		ordered = columnDTOs(cols)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ordered, nil
}

// DeleteColumn removes a column and everything inside it.
func (s *Service) DeleteColumn(ctx context.Context, id uuid.UUID) error {
	affected, err := s.q.DeleteColumn(ctx, id)
	if err != nil {
		return err
	}
	if affected == 0 {
		return notFound("column not found")
	}
	return nil
}

// ---- cards ----------------------------------------------------------------

// CreateCard appends a card to the end of a column.
func (s *Service) CreateCard(ctx context.Context, columnID uuid.UUID, req CardRequest) (Card, error) {
	fields, err := validateCardFields(req)
	if err != nil {
		return Card{}, err
	}

	// A card must hang off a real column; GetColumn gives us the 404.
	if _, err := s.q.GetColumn(ctx, columnID); err != nil {
		return Card{}, mapNotFound(err, "column")
	}

	due, err := dueDateIn(fields.dueDate)
	if err != nil {
		return Card{}, err
	}

	card, err := s.q.CreateCard(ctx, store.CreateCardParams{
		ColumnID:    columnID,
		Title:       fields.title,
		Description: fields.description,
		Assignee:    fields.assignee,
		Priority:    fields.priority,
		DueDate:     due,
	})
	if err != nil {
		return Card{}, err
	}
	return cardDTO(card), nil
}

// UpdateCard applies a full replace of a card's editable fields. The column
// and the ordering are not editable here — moving is its own operation.
func (s *Service) UpdateCard(ctx context.Context, id uuid.UUID, req CardRequest) (Card, error) {
	fields, err := validateCardFields(req)
	if err != nil {
		return Card{}, err
	}

	due, err := dueDateIn(fields.dueDate)
	if err != nil {
		return Card{}, err
	}

	card, err := s.q.UpdateCard(ctx, store.UpdateCardParams{
		ID:          id,
		Title:       fields.title,
		Description: fields.description,
		Assignee:    fields.assignee,
		Priority:    fields.priority,
		DueDate:     due,
	})
	if err != nil {
		return Card{}, mapNotFound(err, "card")
	}
	return cardDTO(card), nil
}

// MoveCard moves a card to a destination column and index.
//
// `position` is the index within the destination column *after* the move. The
// reindex runs in one transaction across both columns, so a failure part way
// through cannot leave a column with duplicate or gapped positions.
func (s *Service) MoveCard(ctx context.Context, id uuid.UUID, columnID uuid.UUID, position int32) (Card, error) {
	var moved Card

	err := s.withTx(ctx, func(q *store.Queries) error {
		card, err := q.GetCard(ctx, id)
		if err != nil {
			return mapNotFound(err, "card")
		}

		if _, err := q.GetColumn(ctx, columnID); err != nil {
			return mapNotFound(err, "column")
		}

		source := card.ColumnID

		if source == columnID {
			// Reorder in place: the card is already in the destination list.
			ids, err := q.ListCardIDsInColumn(ctx, columnID)
			if err != nil {
				return err
			}
			next := insertAt(removeID(ids, id), clampIndex(position, int32(len(ids)-1)), id)
			if err := writePositions(ctx, q, next, id, columnID, false); err != nil {
				return err
			}
		} else {
			sourceIDs, err := q.ListCardIDsInColumn(ctx, source)
			if err != nil {
				return err
			}
			destIDs, err := q.ListCardIDsInColumn(ctx, columnID)
			if err != nil {
				return err
			}

			nextSource := removeID(sourceIDs, id)
			nextDest := insertAt(destIDs, clampIndex(position, int32(len(destIDs))), id)

			if err := writePositions(ctx, q, nextSource, id, source, false); err != nil {
				return err
			}
			if err := writePositions(ctx, q, nextDest, id, columnID, true); err != nil {
				return err
			}
		}

		settled, err := q.GetCard(ctx, id)
		if err != nil {
			return err
		}
		moved = cardDTO(settled)
		return nil
	})
	if err != nil {
		return Card{}, err
	}
	return moved, nil
}

// DeleteCard removes a single card.
func (s *Service) DeleteCard(ctx context.Context, id uuid.UUID) error {
	affected, err := s.q.DeleteCard(ctx, id)
	if err != nil {
		return err
	}
	if affected == 0 {
		return notFound("card not found")
	}
	return nil
}

// Health reports whether the database is reachable.
func (s *Service) Health(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// ---- internals ------------------------------------------------------------

// writePositions rewrites the gap-free ordering of one column.
//
// When reassign is true the card is first moved into this column, so the
// positions it is about to receive are consistent with its new parent.
func writePositions(
	ctx context.Context,
	q *store.Queries,
	ids []uuid.UUID,
	cardID uuid.UUID,
	columnID uuid.UUID,
	reassign bool,
) error {
	if reassign {
		err := q.SetCardColumn(ctx, store.SetCardColumnParams{
			ID:       cardID,
			ColumnID: columnID,
		})
		if err != nil {
			return err
		}
	}

	for i, id := range ids {
		err := q.SetCardPosition(ctx, store.SetCardPositionParams{
			ID:       id,
			Position: int32(i),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// withTx runs fn inside a transaction, rolling back on error.
func (s *Service) withTx(ctx context.Context, fn func(*store.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	// A Rollback after Commit is a harmless no-op.
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(s.q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// cardFields is a validated, normalised CardRequest.
type cardFields struct {
	title       string
	description *string
	assignee    *string
	priority    string
	dueDate     *string
}

// validateCardFields trims and validates a create/update card body. An empty
// description or assignee becomes "" and an empty priority becomes the default.
func validateCardFields(req CardRequest) (cardFields, error) {
	title, err := requireName(req.Title, "title")
	if err != nil {
		return cardFields{}, err
	}

	priority := req.Priority
	if priority == "" {
		priority = defaultPriority
	}
	if _, ok := validPriorities[priority]; !ok {
		return cardFields{}, invalid("priority must be one of low, medium, high, urgent")
	}

	return cardFields{
		title:       title,
		description: normalise(req.Description),
		assignee:    normalise(req.Assignee),
		priority:    priority,
		dueDate:     req.DueDate,
	}, nil
}

// normalise turns nil and whitespace-only text into an empty string, so the
// database never stores "   " as a meaningful assignee.
func normalise(s *string) *string {
	empty := ""
	if s == nil {
		return &empty
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return &empty
	}
	return &trimmed
}

// mapNotFound turns a pgx "no rows" error into the right 404, leaving every
// other error untouched.
func mapNotFound(err error, what string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return notFound("%s not found", what)
	}
	return err
}

// removeID returns ids without the given id. The input is not mutated.
func removeID(ids []uuid.UUID, id uuid.UUID) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(ids))
	for _, existing := range ids {
		if existing != id {
			out = append(out, existing)
		}
	}
	return out
}

// insertAt returns ids with id inserted at idx, appending when idx is at or
// past the end.
func insertAt(ids []uuid.UUID, idx int32, id uuid.UUID) []uuid.UUID {
	idx = clampIndex(idx, int32(len(ids)))
	out := make([]uuid.UUID, 0, len(ids)+1)
	out = append(out, ids[:idx]...)
	out = append(out, id)
	out = append(out, ids[idx:]...)
	return out
}
