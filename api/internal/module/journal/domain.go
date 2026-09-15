// Package journal contains the "journal" business module: domain entities,
// DTOs, ports (Repository & Service), service logic, and the HTTP handler.
package journal

import "time"

// JournalEntry is the domain entity representing a row in `journal_entries`.
// This is the header of a journal transaction (date + description).
type JournalEntry struct {
	ID          int64                // Primary key
	UserID      int64                // FK to users.id
	Date        time.Time            // Transaction date
	Description string               // Optional description
	Lines       []JournalEntryLine   // Associated lines (loaded separately or eagerly)
	CreatedAt   time.Time            // Row creation time
	UpdatedAt   time.Time            // Last modification time
}

// JournalEntryLine is the domain entity representing a row in `journal_entry_lines`.
// Each line represents a debit or credit to a specific account.
// A valid journal entry must have at least 2 lines with balanced total debit = total credit.
type JournalEntryLine struct {
	ID              int64     // Primary key
	JournalEntryID  int64     // FK to journal_entries.id
	AccountID       int64     // FK to accounts.id
	Debit           float64   // Debit amount (must be > 0 if this is a debit line)
	Credit          float64   // Credit amount (must be > 0 if this is a credit line)
	CreatedAt       time.Time // Row creation time
}

// IsDebit returns true if this is a debit line (debit > 0).
func (l *JournalEntryLine) IsDebit() bool {
	return l.Debit > 0
}

// IsCredit returns true if this is a credit line (credit > 0).
func (l *JournalEntryLine) IsCredit() bool {
	return l.Credit > 0
}

// TotalDebit returns the sum of all debit lines in a journal entry.
func TotalDebit(lines []JournalEntryLine) float64 {
	total := 0.0
	for _, l := range lines {
		total += l.Debit
	}
	return total
}

// TotalCredit returns the sum of all credit lines in a journal entry.
func TotalCredit(lines []JournalEntryLine) float64 {
	total := 0.0
	for _, l := range lines {
		total += l.Credit
	}
	return total
}

// IsBalanced returns true if total debit equals total credit.
func IsBalanced(lines []JournalEntryLine) bool {
	return TotalDebit(lines) == TotalCredit(lines)
}

// JSONB is a wrapper for JSONB columns (maps to PostgreSQL JSONB type).
// Mirrors the account module's type to keep modules independent.
type JSONB map[string]interface{}

// JournalAuditLog is the domain entity representing a row in `journal_audit_logs`.
// It records all changes to journal entries (create, update, delete) for audit trail.
type JournalAuditLog struct {
	ID             int64     // Primary key
	UserID         int64     // FK to users.id
	JournalEntryID *int64    // FK to journal_entries.id (nullable — NULL if entry deleted)
	Action         string    // Action name: "journal.created", "journal.updated", "journal.deleted", "journal.bulk_created"
	OldValues      *JSONB    // State before change (nullable)
	NewValues      *JSONB    // State after change (nullable)
	CreatedAt      time.Time // Log creation time
}

// JournalAuditAction represents the possible audit actions for journals.
type JournalAuditAction string

const (
	JournalActionCreated     JournalAuditAction = "journal.created"
	JournalActionUpdated     JournalAuditAction = "journal.updated"
	JournalActionDeleted     JournalAuditAction = "journal.deleted"
	JournalActionBulkCreated JournalAuditAction = "journal.bulk_created"
)

// LinesSnapshot returns a JSON-serializable snapshot of journal lines for audit logs.
func LinesSnapshot(lines []JournalEntryLine) []map[string]interface{} {
	snap := make([]map[string]interface{}, len(lines))
	for i, l := range lines {
		snap[i] = map[string]interface{}{
			"account_id": l.AccountID,
			"debit":      l.Debit,
			"credit":     l.Credit,
		}
	}
	return snap
}
