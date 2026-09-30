package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// demoColumns mirrors the default workflow stages a new board gets. It is
// duplicated from the api package on purpose: store must not import api,
// because api imports store.
var demoColumns = []string{"To Do", "In Progress", "Review", "Done"}

// SeedDemoData inserts a demo board with sample cards and a second, empty
// board — but only when the database has no boards at all.
//
// The emptiness check is what makes this safe to run on every boot: a real
// user's data is never touched, and the seed never runs twice.
func SeedDemoData(ctx context.Context, pool *pgxpool.Pool) error {
	q := New(pool)

	existing, err := q.ListBoards(ctx)
	if err != nil {
		return fmt.Errorf("seed: list boards: %w", err)
	}
	if len(existing) > 0 {
		return nil
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("seed: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tq := q.WithTx(tx)

	roadmap, err := tq.CreateBoard(ctx, "Product Roadmap")
	if err != nil {
		return fmt.Errorf("seed: create demo board: %w", err)
	}

	cols := make(map[string]uuid.UUID, len(demoColumns))
	for _, name := range demoColumns {
		col, err := tq.CreateColumn(ctx, CreateColumnParams{BoardID: roadmap.ID, Name: name})
		if err != nil {
			return fmt.Errorf("seed: create column %q: %w", name, err)
		}
		cols[name] = col.ID
	}

	// Due dates are relative to "now" so the demo always has one overdue
	// card and one comfortably upcoming, whatever day it is run on.
	today := time.Now()
	day := func(offset int) pgtype.Date {
		return pgtype.Date{Time: today.AddDate(0, 0, offset), Valid: true}
	}

	cards := []struct {
		column   string
		title    string
		desc     string
		assignee string
		priority string
		dueDate  pgtype.Date
	}{
		{"To Do", "Redesign the onboarding flow",
			"Cut the current five-step signup down to two. The drop-off analytics show 61% of signups never reach step three.",
			"Priya", "high", day(4)},
		{"To Do", "Add bulk CSV import for boards",
			"Users with 200+ cards currently have to create them one at a time.",
			"", "low", pgtype.Date{}},
		{"In Progress", "Drag-and-drop reordering",
			"Implement smooth card and column reordering with optimistic updates and rollback on failure.",
			"Erik", "urgent", day(1)},
		{"In Progress", "Dark mode polish",
			"Even out contrast in the column surfaces and make the priority badges readable in both themes.",
			"Priya", "medium", day(6)},
		{"Review", "Column task counts",
			"Show a live count per column so WIP is visible at a glance.",
			"Marco", "medium", pgtype.Date{}},
		{"Review", "Keyboard accessibility pass",
			"Verify every card can be moved without a mouse.",
			"Erik", "high", day(-2)},
		{"Done", "Board and column CRUD API",
			"Create, rename, reorder and delete boards and columns.",
			"Marco", "medium", pgtype.Date{}},
		{"Done", "Set up the Go + Postgres backend",
			"sqlc-generated store, migrations, and a seed script.",
			"Erik", "high", pgtype.Date{}},
	}

	for _, c := range cards {
		desc := c.desc
		assignee := c.assignee
		_, err := tq.CreateCard(ctx, CreateCardParams{
			ColumnID:    cols[c.column],
			Title:       c.title,
			Description: &desc,
			Assignee:    &assignee,
			Priority:    c.priority,
			DueDate:     c.dueDate,
		})
		if err != nil {
			return fmt.Errorf("seed: create card %q: %w", c.title, err)
		}
	}

	if _, err := tq.CreateBoard(ctx, "Sales Pipeline"); err != nil {
		return fmt.Errorf("seed: create second board: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("seed: commit: %w", err)
	}
	return nil
}
