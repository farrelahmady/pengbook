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
