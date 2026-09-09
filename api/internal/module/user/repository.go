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
	Create(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByEmailOrUsername(ctx context.Context, identifier string) (*User, error)
	InsertAuditLog(ctx context.Context, audit *AuditLog) error
}
