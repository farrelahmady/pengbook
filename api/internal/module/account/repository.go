// Package account contains the "account" business module: domain entities,
// DTOs, ports (Repository & Service), service logic, and the HTTP handler.
package account

import "context"

// Repository is the PORT (interface) for account data access.
// Services depend only on this interface, not on the concrete implementation.
type Repository interface {
	// Create stores a new account. ID and timestamps are filled from the DB.
	Create(ctx context.Context, a *Account) error

	// FindByID returns the account for the given id.
	// Returns (nil, nil) when not found.
	FindByID(ctx context.Context, id int64) (*Account, error)

	// FindByUserID returns all accounts for the given user, ordered by code.
	FindByUserID(ctx context.Context, userID int64) ([]Account, error)

	// Update updates an existing account.
	Update(ctx context.Context, a *Account) error

	// Delete deletes an account by id.
	Delete(ctx context.Context, id int64) error

	// CountByUserID returns the count of accounts for the given user.
	CountByUserID(ctx context.Context, userID int64) (int64, error)

	// CountPostingByUserID returns the count of posting accounts (level=3) for the given user.
	CountPostingByUserID(ctx context.Context, userID int64) (int64, error)

	// CountHeaderByUserID returns the count of header accounts (level<3) for the given user.
	CountHeaderByUserID(ctx context.Context, userID int64) (int64, error)

	// InsertAuditLog records an account action into the audit table.
	InsertAuditLog(ctx context.Context, log *AccountAuditLog) error

	// FindPostingByUserID returns all posting accounts (level=3) for the given user.
	FindPostingByUserID(ctx context.Context, userID int64) ([]Account, error)
}
