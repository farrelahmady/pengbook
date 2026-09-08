package user

import "context"

// Repository is the PORT (interface) for user data access.
//
// Services depend only on this interface, not on the concrete implementation
// (pgx). The implementation lives in internal/infrastructure/postgres.
// Every method takes a ctx so the repository can read the transaction from
// the context (via database.GetTx): a transaction exists → use it, otherwise
// → use the regular connection.
type Repository interface {
	// Create stores a new user. ID and timestamps are filled from the DB
	// (RETURNING clause).
	Create(ctx context.Context, u *User) error

	// FindByID returns the user for the given id.
	// Returns (nil, nil) when not found.
	FindByID(ctx context.Context, id int64) (*User, error)

	// FindByEmail returns the user for the given email.
	// Returns (nil, nil) when not found.
	FindByEmail(ctx context.Context, email string) (*User, error)

	// InsertAuditLog records a user action into the audit table.
	InsertAuditLog(ctx context.Context, audit *AuditLog) error
}
