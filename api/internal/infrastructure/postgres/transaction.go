// Package postgres contains the infrastructure implementation using the
// PostgreSQL driver pgx/v5: repositories and the transaction manager (TxManager).
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengbook/api/internal/database"
)

// txManager is the database.TxManager implementation backed by pgxpool.
//
// A single instance for the whole app (injected from cmd/api/main.go), so all
// services share the same transaction logic — no per-service WithTransaction.
type txManager struct {
	pool *pgxpool.Pool // used to Begin new transactions
}

// NewTxManager creates the shared transaction manager.
func NewTxManager(pool *pgxpool.Pool) database.TxManager {
	return &txManager{pool: pool}
}

// WithTransaction runs fn inside a transaction.
//
// Two paths:
//  1. If a transaction already exists in the context (GetTx non-nil — e.g.
//     injected by a test), fn runs as-is WITHOUT begin/commit. The caller
//     controls commit or rollback (a test may roll back at the end).
//  2. Otherwise: Begin a new transaction → run fn with the transaction in the
//     context → Commit on success, Rollback when fn errors/panics (defer
//     rollback guarantees safety even on panic).
//
// This is what allows transactions to be "hijacked" from a test file: the test
// begins its own transaction, injects it into the context, and rolls back at
// the end — the service never commits.
func (m *txManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if database.GetTx(ctx) != nil {
		return fn(ctx)
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(database.WithTx(ctx, tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
