package postgres

import (
	"context"
	"time"

	"pengbook/api/internal/module/account"
	"pengbook/api/internal/module/journal"
)

// journalMutatorAdapter implements account.JournalMutator by wrapping
// the journal.Repository interface, doing type conversion between packages.
type journalMutatorAdapter struct {
	repo journal.Repository
}

// NewJournalMutatorAdapter creates an account.JournalMutator that wraps
// the journal.Repository.
func NewJournalMutatorAdapter(repo journal.Repository) account.JournalMutator {
	return &journalMutatorAdapter{repo: repo}
}

func (a *journalMutatorAdapter) CreateEntry(ctx context.Context, userID int64, date time.Time, description string, lines []account.JournalEntryLineInput) (int64, error) {
	// Convert account.JournalEntryLineInput to journal.JournalEntryLine
	journalLines := make([]journal.JournalEntryLine, len(lines))
	for i, l := range lines {
		journalLines[i] = journal.JournalEntryLine{
			AccountID: l.AccountID,
			Debit:     l.Debit,
			Credit:    l.Credit,
		}
	}

	entry := &journal.JournalEntry{
		UserID:      userID,
		Date:        date,
		Description: description,
		Lines:       journalLines,
	}

	if err := a.repo.CreateEntry(ctx, entry); err != nil {
		return 0, err
	}
	return entry.ID, nil
}

func (a *journalMutatorAdapter) ApplyBalanceDelta(ctx context.Context, deltas map[int64]float64) error {
	return a.repo.ApplyBalanceDelta(ctx, deltas)
}

func (a *journalMutatorAdapter) InsertAuditLog(ctx context.Context, userID int64, journalEntryID *int64, action string, newValues account.JSONB) error {
	// Convert account.JSONB to journal.JSONB
	journalNewValues := journal.JSONB{}
	for k, v := range newValues {
		journalNewValues[k] = v
	}

	log := &journal.JournalAuditLog{
		UserID:         userID,
		JournalEntryID: journalEntryID,
		Action:         action,
		NewValues:      &journalNewValues,
	}
	return a.repo.InsertAuditLog(ctx, log)
}
