package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// demoColumns mirrors the default workflow stages a new board gets. It is
// duplicated from the api package on purpose: store must not import api,
// because api imports store.
var demoColumns = []string{"To Do", "In Progress", "Review", "Done"}

// demoCards is the sample content each persona receives, so logging in as a
// different account shows visibly different boards rather than an empty shell.
//
// Assignees are free text naming people involved in the work — deliberately not
// account references, since boards are owner-only and nobody else can open
// these cards.
var demoCards = []struct {
	column   string
	title    string
	desc     string
	assignee string
	priority string
	dueIn    int
}{
	{"To Do", "Redesign the onboarding flow",
		"Cut the current five-step signup down to two. Drop-off analytics show 61% of signups never reach step three.",
		"Priya", "high", 4},
	{"To Do", "Add bulk CSV import for boards",
		"Users with 200+ cards currently have to create them one at a time.",
		"", "low", 0},
	{"In Progress", "Drag-and-drop reordering",
		"Smooth card and column reordering with optimistic updates and rollback on failure.",
		"Erik", "urgent", 1},
	{"In Progress", "Dark mode polish",
		"Even out contrast in the column surfaces and make priority badges readable in both themes.",
		"Priya", "medium", 6},
	{"Review", "Column task counts",
		"Show a live count per column so WIP is visible at a glance.",
		"Marco", "medium", 0},
	{"Review", "Keyboard accessibility pass",
		"Verify every card can be moved without a mouse.",
		"Erik", "high", -2},
	{"Done", "Board and column CRUD API",
		"Create, rename, reorder and delete boards and columns.",
		"Marco", "medium", 0},
	{"Done", "Set up the Go + Postgres backend",
		"sqlc-generated store, migrations, and a seed script.",
		"Erik", "high", 0},
}

// personaBoardNames give each demo persona a differently-named board, so it is
// obvious at a glance which account you are looking at.
var personaBoardNames = map[string]string{
	"alice": "Product Roadmap",
	"bob":   "Sales Pipeline",
	"carol": "Marketing Launch",
	"demo":  "Getting Started",
}

// SeedResult reports what the seed actually did, so boot can log honestly
// rather than claiming a seed that was skipped.
type SeedResult struct {
	CreatedPersonas []string
	CreatedBoards   int
}

// SeedDemoData provisions the demo accounts and gives each one sample boards.
//
// The emptiness check is what makes this safe to run on every boot: accounts
// and boards are only created when the database has none, so it never touches
// real data and never runs twice.
//
// `legacy` is handled specially. It is created by migration 00002 and owns any
// boards that predate ownership, so it is never given demo content — it is the
// archive account, not a persona.
func SeedDemoData(
	ctx context.Context,
	pool *pgxpool.Pool,
	personas []string,
	demoPassword string,
) (SeedResult, error) {
	var result SeedResult

	q := New(pool)

	userCount, err := q.CountUsers(ctx)
	if err != nil {
		return result, fmt.Errorf("seed: count users: %w", err)
	}

	// Hash once: every persona shares the same password, so there is no reason
	// to pay bcrypt's cost four times.
	hash, err := hashPassword(demoPassword)
	if err != nil {
		return result, fmt.Errorf("seed: hash demo password: %w", err)
	}

	for _, username := range personas {
		existing, err := q.GetUserByUsername(ctx, username)
		if err == nil {
			// The account already exists (typically `legacy`, or a previous
			// seed). Keep the stored hash but refresh it from the current
			// DEMO_PASSWORD so changing the env var still takes effect.
			if err := q.UpdateUserPassword(ctx, UpdateUserPasswordParams{
				ID: existing.ID, PasswordHash: hash,
			}); err != nil {
				return result, fmt.Errorf("seed: refresh password for %q: %w", username, err)
			}
			continue
		}
		if !isNoRows(err) {
			return result, fmt.Errorf("seed: look up %q: %w", username, err)
		}

		user, err := q.UpsertPersona(ctx, UpsertPersonaParams{
			Username: username, PasswordHash: hash,
		})
		if err != nil {
			return result, fmt.Errorf("seed: create persona %q: %w", username, err)
		}
		result.CreatedPersonas = append(result.CreatedPersonas, username)

		boards, err := seedBoardsFor(ctx, pool, user.ID, personaBoardNames[username])
		if err != nil {
			return result, fmt.Errorf("seed: boards for %q: %w", username, err)
		}
		result.CreatedBoards += boards
	}

	// Only skip content creation when the database already had accounts. If we
	// just created them, an empty board count means something is wrong.
	if userCount > 0 && result.CreatedBoards == 0 && len(result.CreatedPersonas) == 0 {
		return result, nil
	}

	return result, nil
}

// seedBoardsFor gives a new persona one populated board and one empty one, so
// both the populated and empty states are visible without deleting anything.
func seedBoardsFor(ctx context.Context, pool *pgxpool.Pool, owner uuid.UUID, name string) (int, error) {
	if name == "" {
		name = "My Board"
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tq := New(pool).WithTx(tx)

	board, err := tq.CreateBoard(ctx, CreateBoardParams{Name: name, OwnerID: owner})
	if err != nil {
		return 0, err
	}

	cols := make(map[string]uuid.UUID, len(demoColumns))
	for _, stage := range demoColumns {
		col, err := tq.CreateColumn(ctx, CreateColumnParams{BoardID: board.ID, Name: stage})
		if err != nil {
			return 0, err
		}
		cols[stage] = col.ID
	}

	// Due dates are relative to now so the demo always has one overdue card and
	// one comfortably upcoming, whatever day it is run on.
	today := time.Now()

	for _, c := range demoCards {
		desc := c.desc
		assignee := c.assignee
		_, err := tq.CreateCard(ctx, CreateCardParams{
			ColumnID:    cols[c.column],
			Title:       c.title,
			Description: &desc,
			Assignee:    &assignee,
			Priority:    c.priority,
			DueDate:     dueOffset(today, c.dueIn),
		})
		if err != nil {
			return 0, err
		}
	}

	// A second, empty board shows the empty state without deleting the sample
	// content above.
	if _, err := tq.CreateBoard(ctx, CreateBoardParams{
		Name:    name + " (empty)",
		OwnerID: owner,
	}); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return 2, nil
}
