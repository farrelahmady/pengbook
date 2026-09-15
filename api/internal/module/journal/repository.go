// Package journal contains the "journal" business module: domain entities,
// DTOs, ports (Repository & Service), service logic, and the HTTP handler.
package journal

import (
	"context"
	"time"
)

// Repository is the PORT (interface) for journal data access.
type Repository interface {
	// CreateEntry creates a new journal entry with its lines in a transaction.
	CreateEntry(ctx context.Context, entry *JournalEntry) error

	// CreateEntries creates multiple journal entries with their lines in a batch.
	CreateEntries(ctx context.Context, entries []JournalEntry) error

	// FindEntryByID returns a journal entry by id, including its lines.
	FindEntryByID(ctx context.Context, id int64) (*JournalEntry, error)

	// FindEntriesByUserID returns paginated journal entries for a user.
	FindEntriesByUserID(ctx context.Context, userID int64, filter EntryFilter) ([]JournalEntry, error)

	// FindEntriesByUserIDWithNetEffect returns journal entries with net effect calculation (for list view).
	FindEntriesByUserIDWithNetEffect(ctx context.Context, userID int64, filter EntryFilter) ([]JournalEntryListItem, error)

	// UpdateEntry updates a journal entry and replaces its lines.
	UpdateEntry(ctx context.Context, entry *JournalEntry) error

	// DeleteEntry deletes a journal entry and its lines.
	DeleteEntry(ctx context.Context, id int64) error

	// GetSummaryByUserID returns aggregated summary (total debit, total credit, count) in a single query.
	GetSummaryByUserID(ctx context.Context, userID int64) (*JournalSummary, error)

	// CountByUserID returns the total count of journal entries for a user.
	CountByUserID(ctx context.Context, userID int64) (int64, error)

	// SumDebitByUserID returns the total debit for a user.
	SumDebitByUserID(ctx context.Context, userID int64) (float64, error)

	// SumCreditByUserID returns the total credit for a user.
	SumCreditByUserID(ctx context.Context, userID int64) (float64, error)

	// CountByUserIDWithFilter returns the count of journal entries with filter.
	CountByUserIDWithFilter(ctx context.Context, userID int64, filter EntryFilter) (int64, error)

	// SumDebitByUserIDWithFilter returns the total debit with filter.
	SumDebitByUserIDWithFilter(ctx context.Context, userID int64, filter EntryFilter) (float64, error)

	// SumCreditByUserIDWithFilter returns the total credit with filter.
	SumCreditByUserIDWithFilter(ctx context.Context, userID int64, filter EntryFilter) (float64, error)

	// RecalculateAccountBalance recalculates account balances for the given accounts or all accounts for a user.
	// Kept as the healing/reconciliation path. Hot path (Create/Update/CreateBulk)
	// uses ApplyBalanceDelta instead (O(lines in entry), not O(all lines)).
	RecalculateAccountBalance(ctx context.Context, accountIDs []int64, userID int64) error

	// ApplyBalanceDelta applies pre-signed balance deltas atomically.
	// deltas maps account_id -> signed amount, where signed amount is
	// (debit - credit) * signFactor. signFactor is +1 for ASSET/EXPENSE/OTHER
	// and -1 for LIABILITY/EQUITY/REVENUE (same as recalculate_account_balance).
	// Implemented as a single UPSERT statement so concurrent writers cannot
	// lose updates, and accounts without a balance row get one created.
	ApplyBalanceDelta(ctx context.Context, deltas map[int64]float64) error
}

// EntryFilter contains filter options for listing journal entries.
type EntryFilter struct {
	Limit           int        `json:"limit"`
	CursorDatetime  *time.Time `json:"cursor_datetime,omitempty"`
	CursorID        *int64     `json:"cursor_id,omitempty"`
	StartDate       *time.Time `json:"start_date,omitempty"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	AccountIDs      []int64    `json:"account_ids,omitempty"`
}
