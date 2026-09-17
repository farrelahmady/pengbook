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

	// FindByIDs returns accounts for the given ids in a single query.
	// Returns a map of id -> Account for quick lookup.
	FindByIDs(ctx context.Context, ids []int64) (map[int64]*Account, error)

	// FindByCodes returns accounts for the given codes in a single query.
	// Returns a map of code -> Account for quick lookup.
	FindByCodes(ctx context.Context, codes []string) (map[string]*Account, error)

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

	// FindByLevelAndType returns all accounts of the given user at the given level and type if exists,
	// ordered by code. Used to list candidate parents for account creation.
	FindByLevelAndType(ctx context.Context, userID int64, vel int8, typeAccount string) ([]Account, error)

	// FindMaxChildCode returns the highest child code of the given parent.
	// Returns ("", false, nil) when the parent has no children yet.
	FindMaxChildCode(ctx context.Context, userID int64, parentID int64) (string, bool, error)

	// SeedRoots creates the six fixed level-0 roots for a new user.
	// Idempotent: existing roots are left untouched.
	SeedRoots(ctx context.Context, userID int64) error

	// EnsureBalance creates the zero balance row for an Asset posting account.
	// Idempotent: an existing row is left untouched.
	EnsureBalance(ctx context.Context, accountID int64) error

	// FindAssetWithBalances returns all ASSET accounts of the given user
	// with their cached balance (COALESCE to 0 when the balance row is
	// missing), ordered by code. Used by the Asset page (summary + groups).
	FindAssetWithBalances(ctx context.Context, userID int64) ([]AssetBalanceRow, error)
}

// AssetBalanceRow is one ASSET account row joined with its cached balance.
type AssetBalanceRow struct {
	Account Account
	Balance float64
}
