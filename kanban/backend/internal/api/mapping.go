package api

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"kanban/internal/store"
)

// dateLayout is the wire format for cards.due_date.
//
// A due date is a calendar day, not an instant. It is formatted in UTC from
// the pgtype.Date value, which Postgres hands us as UTC midnight, so a client
// in any timezone receives exactly the day that was stored. The frontend must
// format it locally for display and must never round-trip it through
// new Date(...).toISOString(), which would shift it by a day west of UTC.
const dateLayout = "2006-01-02"

func boardDTO(b store.Board) Board {
	return Board{
		ID:        b.ID.String(),
		Name:      b.Name,
		Position:  b.Position,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

func columnDTO(c store.BoardColumn) Column {
	return Column{
		ID:        c.ID.String(),
		BoardID:   c.BoardID.String(),
		Name:      c.Name,
		Position:  c.Position,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func cardDTO(c store.Card) Card {
	return Card{
		ID:          c.ID.String(),
		ColumnID:    c.ColumnID.String(),
		Title:       c.Title,
		Description: c.Description,
		Assignee:    c.Assignee,
		Priority:    c.Priority,
		DueDate:     dueDateOut(c.DueDate),
		Position:    c.Position,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// dueDateOut converts a nullable pgtype.Date to a nullable "YYYY-MM-DD".
func dueDateOut(d pgtype.Date) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.UTC().Format(dateLayout)
	return &s
}

// dueDateIn converts a nullable "YYYY-MM-DD" to a pgtype.Date. An absent or
// empty value means "no due date", which is the zero pgtype.Date.
func dueDateIn(s *string) (pgtype.Date, error) {
	if s == nil || *s == "" {
		return pgtype.Date{}, nil
	}
	t, err := time.Parse(dateLayout, *s)
	if err != nil {
		return pgtype.Date{}, invalid("due_date must be formatted as YYYY-MM-DD")
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func boardDTOs(bs []store.Board) []Board {
	out := make([]Board, 0, len(bs))
	for _, b := range bs {
		out = append(out, boardDTO(b))
	}
	return out
}

func columnDTOs(cs []store.BoardColumn) []Column {
	out := make([]Column, 0, len(cs))
	for _, c := range cs {
		out = append(out, columnDTO(c))
	}
	return out
}

func cardDTOs(cs []store.Card) []Card {
	out := make([]Card, 0, len(cs))
	for _, c := range cs {
		out = append(out, cardDTO(c))
	}
	return out
}
