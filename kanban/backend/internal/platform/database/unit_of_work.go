// Package database holds connection-pool concerns and the transaction helper
// shared by feature slices.
package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"kanban/internal/store"
)

// UnitOfWork runs fn inside a single transaction.
//
// This exists because a vertical slice owns one operation, but some operations
// are not confined to one resource: moving a card rewrites positions in both a
// source and a destination column. Slices cannot each own a private
// transaction helper without either duplicating it or importing each other, so
// it lives here as platform code.
//
// The callback receives a *store.Queries bound to the transaction. Returning
// an error rolls back; returning nil commits.
func UnitOfWork(ctx context.Context, pool *pgxpool.Pool, fn func(q *store.Queries) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	// Rollback after Commit is a harmless no-op.
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(store.New(pool).WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// IsNotFound reports whether err is pgx's "no rows in result set".
//
// It is the bridge between sqlc's `:one` queries returning ErrNoRows and the
// domain layer naming what was missing.
func IsNotFound(err error) bool {
	return err != nil && err == pgx.ErrNoRows
}
