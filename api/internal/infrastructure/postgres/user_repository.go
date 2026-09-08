package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pengbook/api/internal/database"
	"pengbook/api/internal/module/user"
)

// userRepository is the pgx implementation of the user.Repository port.
type userRepository struct {
	pool *pgxpool.Pool // connection pool (used when no transaction exists)
}

// NewUserRepository creates the Repository implementation for user.
func NewUserRepository(pool *pgxpool.Pool) user.Repository {
	return &userRepository{pool: pool}
}

// db selects the query target based on the context:
//
//   - If a transaction exists in the context (database.GetTx non-nil) → use it.
//   - Otherwise → use the regular connection pool.
//
// This is the core of the "use the transaction if present, otherwise the
// regular connection" pattern. Every query method below goes through it.
func (r *userRepository) db(ctx context.Context) database.DBTX {
	if tx := database.GetTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

// createUserQuery: INSERT + RETURNING so id/timestamps are filled by the DB.
const createUserQuery = `
	INSERT INTO users (name, email, password_hash)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at
`

// Create stores a new user. u.ID, u.CreatedAt, u.UpdatedAt are filled from DB.
func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	return r.db(ctx).QueryRow(ctx, createUserQuery, u.Name, u.Email, u.PasswordHash).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// findByIDQuery: full SELECT of a single user row.
const findByIDQuery = `
	SELECT id, name, email, password_hash, created_at, updated_at
	FROM users
	WHERE id = $1
`

// FindByID returns the user for the given id.
// When no row matches, returns (nil, nil) — not an error.
func (r *userRepository) FindByID(ctx context.Context, id int64) (*user.User, error) {
	var u user.User
	err := r.db(ctx).QueryRow(ctx, findByIDQuery, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// findByEmailQuery: full SELECT of a single user row by email.
const findByEmailQuery = `
	SELECT id, name, email, password_hash, created_at, updated_at
	FROM users
	WHERE email = $1
`

// FindByEmail returns the user for the given email.
// When no row matches, returns (nil, nil) — not an error.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.db(ctx).QueryRow(ctx, findByEmailQuery, email).
		Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// insertAuditLogQuery: INSERT audit log + RETURNING to fill ID/timestamp.
const insertAuditLogQuery = `
	INSERT INTO user_audit_logs (user_id, action)
	VALUES ($1, $2)
	RETURNING id, created_at
`

// InsertAuditLog records a user action into the audit table.
func (r *userRepository) InsertAuditLog(ctx context.Context, audit *user.AuditLog) error {
	return r.db(ctx).QueryRow(ctx, insertAuditLogQuery, audit.UserID, audit.Action).
		Scan(&audit.ID, &audit.CreatedAt)
}
