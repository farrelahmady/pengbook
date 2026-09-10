package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pengbook/api/internal/database"
	"pengbook/api/internal/module/account"
	"pengbook/api/pkg/logger"
)

// accountRepository is the pgx implementation of the account.Repository port.
type accountRepository struct {
	pool *pgxpool.Pool
}

// NewAccountRepository creates the Repository implementation for account.
func NewAccountRepository(pool *pgxpool.Pool) account.Repository {
	return &accountRepository{pool: pool}
}

// db selects the query target based on the context.
func (r *accountRepository) db(ctx context.Context) database.DBTX {
	if tx := database.GetTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

const createAccountQuery = `
	INSERT INTO accounts (user_id, code, name, parent_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id, type, level, created_at, updated_at
`

func (r *accountRepository) Create(ctx context.Context, a *account.Account) error {
	log := logger.FromContext(ctx)
	err := r.db(ctx).QueryRow(ctx, createAccountQuery,
		a.UserID, a.Code, a.Name, a.ParentID,
	).Scan(&a.ID, &a.Type, &a.Level, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		log.Error("repo: failed to create account", "user_id", a.UserID, "code", a.Code, "error", err)
	}
	return err
}

const findAccountByIDQuery = `
	SELECT id, user_id, code, name, type, level, parent_id, created_at, updated_at
	FROM accounts
	WHERE id = $1
`

func (r *accountRepository) FindByID(ctx context.Context, id int64) (*account.Account, error) {
	log := logger.FromContext(ctx)
	var a account.Account
	err := r.db(ctx).QueryRow(ctx, findAccountByIDQuery, id).
		Scan(&a.ID, &a.UserID, &a.Code, &a.Name, &a.Type, &a.Level, &a.ParentID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		log.Error("repo: failed to find account by id", "account_id", id, "error", err)
		return nil, err
	}
	return &a, nil
}

const findByUserIDQuery = `
	SELECT id, user_id, code, name, type, level, parent_id, created_at, updated_at
	FROM accounts
	WHERE user_id = $1
	ORDER BY code
`

func (r *accountRepository) FindByUserID(ctx context.Context, userID int64) ([]account.Account, error) {
	rows, err := r.db(ctx).Query(ctx, findByUserIDQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []account.Account
	for rows.Next() {
		var a account.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Code, &a.Name, &a.Type, &a.Level, &a.ParentID, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

const updateAccountQuery = `
	UPDATE accounts
	SET name = $2, parent_id = $3, updated_at = NOW()
	WHERE id = $1
	RETURNING updated_at
`

func (r *accountRepository) Update(ctx context.Context, a *account.Account) error {
	log := logger.FromContext(ctx)
	err := r.db(ctx).QueryRow(ctx, updateAccountQuery,
		a.ID, a.Name, a.ParentID,
	).Scan(&a.UpdatedAt)
	if err != nil {
		log.Error("repo: failed to update account", "account_id", a.ID, "error", err)
	}
	return err
}

const deleteAccountQuery = `DELETE FROM accounts WHERE id = $1`

func (r *accountRepository) Delete(ctx context.Context, id int64) error {
	log := logger.FromContext(ctx)
	_, err := r.db(ctx).Exec(ctx, deleteAccountQuery, id)
	if err != nil {
		log.Error("repo: failed to delete account", "account_id", id, "error", err)
	}
	return err
}

const countByUserIDQuery = `SELECT COUNT(*) FROM accounts WHERE user_id = $1`

func (r *accountRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db(ctx).QueryRow(ctx, countByUserIDQuery, userID).Scan(&count)
	return count, err
}

const countPostingByUserIDQuery = `SELECT COUNT(*) FROM accounts WHERE user_id = $1 AND level = 3`

func (r *accountRepository) CountPostingByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db(ctx).QueryRow(ctx, countPostingByUserIDQuery, userID).Scan(&count)
	return count, err
}

const countHeaderByUserIDQuery = `SELECT COUNT(*) FROM accounts WHERE user_id = $1 AND level < 3`

func (r *accountRepository) CountHeaderByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db(ctx).QueryRow(ctx, countHeaderByUserIDQuery, userID).Scan(&count)
	return count, err
}

const accountInsertAuditLogQuery = `
	INSERT INTO accounts_audit_logs (user_id, account_id, action, old_values, new_values)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at
`

func (r *accountRepository) InsertAuditLog(ctx context.Context, logEntry *account.AccountAuditLog) error {
	log := logger.FromContext(ctx)
	var oldValues, newValues []byte
	if logEntry.OldValues != nil {
		oldValues, _ = json.Marshal(logEntry.OldValues)
	}
	if logEntry.NewValues != nil {
		newValues, _ = json.Marshal(logEntry.NewValues)
	}
	err := r.db(ctx).QueryRow(ctx, accountInsertAuditLogQuery,
		logEntry.UserID, logEntry.AccountID, logEntry.Action, oldValues, newValues,
	).Scan(&logEntry.ID, &logEntry.CreatedAt)
	if err != nil {
		log.Error("repo: failed to insert account audit log", "user_id", logEntry.UserID, "account_id", logEntry.AccountID, "action", logEntry.Action, "error", err)
	}
	return err
}

const findPostingByUserIDQuery = `
	SELECT id, user_id, code, name, type, level, parent_id, created_at, updated_at
	FROM accounts
	WHERE user_id = $1 AND level = 3
	ORDER BY code
`

func (r *accountRepository) FindPostingByUserID(ctx context.Context, userID int64) ([]account.Account, error) {
	rows, err := r.db(ctx).Query(ctx, findPostingByUserIDQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []account.Account
	for rows.Next() {
		var a account.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Code, &a.Name, &a.Type, &a.Level, &a.ParentID, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}
