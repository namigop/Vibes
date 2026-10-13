package domain

import "kanban/internal/store"

// Mapping between sqlc's generated row structs and the domain entities.
//
// The two are kept separate on purpose: the database shape can change without
// silently changing the wire format. Every list is built with make(...) so it
// serialises as [] rather than null — the frontend guards against the latter,
// but an empty array is simply correct.
func BoardFromStore(b store.Board) Board {
	return Board{
		ID:        b.ID.String(),
		Name:      b.Name,
		Position:  b.Position,
		OwnerID:   b.OwnerID.String(),
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

func UserFromStore(u store.User) User {
	return User{ID: u.ID.String(), Username: u.Username}
}

func ColumnFromStore(c store.BoardColumn) Column {
	return Column{
		ID:        c.ID.String(),
		BoardID:   c.BoardID.String(),
		Name:      c.Name,
		Position:  c.Position,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func CardFromStore(c store.Card) Card {
	return Card{
		ID:          c.ID.String(),
		ColumnID:    c.ColumnID.String(),
		Title:       c.Title,
		Description: c.Description,
		Assignee:    c.Assignee,
		Priority:    c.Priority,
		DueDate:     DueDateOut(c.DueDate),
		Position:    c.Position,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func BoardsFromStore(rows []store.Board) []Board {
	out := make([]Board, 0, len(rows))
	for _, b := range rows {
		out = append(out, BoardFromStore(b))
	}
	return out
}

func ColumnsFromStore(rows []store.BoardColumn) []Column {
	out := make([]Column, 0, len(rows))
	for _, c := range rows {
		out = append(out, ColumnFromStore(c))
	}
	return out
}

func CardsFromStore(rows []store.Card) []Card {
	out := make([]Card, 0, len(rows))
	for _, c := range rows {
		out = append(out, CardFromStore(c))
	}
	return out
}
