// Package account contains the "account" business module: domain entities,
// DTOs, ports (Repository & Service), service logic, and the HTTP handler.
package account

import (
	"encoding/json"
	"time"
)

// AccountType represents the type of account, auto-generated from code.
type AccountType string

const (
	AccountTypeAsset      AccountType = "ASSET"
	AccountTypeLiability  AccountType = "LIABILITY"
	AccountTypeEquity     AccountType = "EQUITY"
	AccountTypeRevenue    AccountType = "REVENUE"
	AccountTypeExpense    AccountType = "EXPENSE"
	AccountTypeOther      AccountType = "OTHER"
)

// Account is the domain entity representing a row in `accounts`.
// Type, Level are generated columns from code — read-only from DB.
type Account struct {
	ID        int64       // Primary key
	UserID    int64       // FK to users.id
	Code      string      // Account code, e.g. "1.01.01.01"
	Name      string      // Account name, e.g. "Mandiri - Main"
	Type      AccountType // Generated: ASSET/LIABILITY/EQUITY/REVENUE/EXPENSE/OTHER
	Level     int8        // Generated: 0 (type), 1 (category), 2 (sub-category), 3 (posting)
	ParentID  *int64      // FK to accounts.id (nullable, nil for level 0)
	CreatedAt time.Time   // Row creation time
	UpdatedAt time.Time   // Last modification time
}

// CanPost returns true if this account can have journal entries posted to it.
// Only level 3 accounts (posting accounts) can be used in journal entry lines.
func (a *Account) CanPost() bool {
	return a.Level == 3
}

// AccountBalance is the domain entity representing a row in `account_balances`.
// It stores the cached balance for an account, updated when journal entries are created.
type AccountBalance struct {
	AccountID int64     // Primary key, FK to accounts.id
	Balance   float64   // Current balance (debit - credit)
	UpdatedAt time.Time // Last balance update time
}

// JSONB is a wrapper for JSONB columns (maps to PostgreSQL JSONB type).
type JSONB map[string]interface{}

// AccountAuditLog is the domain entity representing a row in `accounts_audit_logs`.
// It records all changes to accounts (create, update, delete) for audit trail.
type AccountAuditLog struct {
	ID        int64     // Primary key
	UserID    int64     // FK to users.id
	AccountID *int64    // FK to accounts.id (nullable — NULL if account deleted)
	Action    string    // Action name: "account.created", "account.updated", "account.deleted"
	OldValues *JSONB    // State before change (nullable)
	NewValues *JSONB    // State after change (nullable)
	CreatedAt time.Time // Log creation time
}

// AccountAuditAction represents the possible audit actions for accounts.
type AccountAuditAction string

const (
	AccountActionCreated AccountAuditAction = "account.created"
	AccountActionUpdated AccountAuditAction = "account.updated"
	AccountActionDeleted AccountAuditAction = "account.deleted"
)

// ToJSON converts the AccountAuditLog to a JSON-serializable map (for logging/debugging).
func (a *AccountAuditLog) ToJSON() []byte {
	data := map[string]interface{}{
		"id":         a.ID,
		"user_id":    a.UserID,
		"account_id": a.AccountID,
		"action":     a.Action,
		"created_at": a.CreatedAt,
	}
	if a.OldValues != nil {
		data["old_values"] = a.OldValues
	}
	if a.NewValues != nil {
		data["new_values"] = a.NewValues
	}
	b, _ := json.Marshal(data)
	return b
}
