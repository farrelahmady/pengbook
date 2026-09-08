package database

import "context"

// txKey is the unique key used to store a transaction inside a context.
type txKey struct{}

// WithTx stores a transaction (or any DBTX object) into the context so that
// repositories called inside that transaction can retrieve and use it via GetTx.
func WithTx(ctx context.Context, tx DBTX) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// GetTx retrieves the transaction from the context.
//
// Returns nil when no transaction was injected — the repository then falls
// back to the regular connection pool.
func GetTx(ctx context.Context) DBTX {
	tx, _ := ctx.Value(txKey{}).(DBTX)
	return tx
}

// TxManager is the transaction manager contract used by services.
//
// The interface is defined ONCE here (not per-service) and implemented in
// infrastructure/postgres. Services only need to hold a database.TxManager
// and call WithTransaction — no need to write begin/commit/rollback logic.
type TxManager interface {
	// WithTransaction runs fn inside a transaction.
	//
	// If a transaction already exists in the context (e.g. injected by a
	// test), fn runs as-is WITHOUT commit/rollback — the caller manages it.
	// Otherwise a new transaction is started, committed on success, and
	// rolled back if fn errors or panics.
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
