package journal

import (
	"context"
	"errors"
	"time"

	"pengbook/api/internal/database"
	"pengbook/api/internal/module/account"
)

// Errors
var (
	ErrNotFound       = errors.New("journal entry not found")
	ErrNotBalanced    = errors.New("journal entry lines are not balanced (debit != credit)")
	ErrInvalidLines   = errors.New("journal entry must have at least 2 lines")
	ErrNotPostingAccount = errors.New("account is not a posting account (level != 3)")
)

// Service is the PORT (interface) for journal business logic.
type Service interface {
	// GetAllScrollView returns paginated journal entries with filters.
	GetAllScrollView(ctx context.Context, userID int64, filter ListRequest) (*ListResponse, error)

	// GetTotalSummary returns aggregate totals for a user.
	GetTotalSummary(ctx context.Context, userID int64) (*JournalSummary, error)

	// Create creates a new journal entry with balanced lines.
	Create(ctx context.Context, userID int64, req CreateJournalRequest) (*JournalEntryResponse, error)

	// Update updates an existing journal entry.
	Update(ctx context.Context, userID int64, entryID int64, req UpdateJournalRequest) (*JournalEntryResponse, error)
}

type service struct {
	repo      Repository
	accountRepo account.Repository
	tx        database.TxManager
}

func NewService(repo Repository, accountRepo account.Repository, tx database.TxManager) Service {
	return &service{repo: repo, accountRepo: accountRepo, tx: tx}
}

func (s *service) GetAllScrollView(ctx context.Context, userID int64, filter ListRequest) (*ListResponse, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	entryFilter := EntryFilter{
		Page:       page,
		Limit:      limit,
		AccountIDs: filter.AccountIDs,
	}

	if filter.StartDate != "" {
		t, err := time.Parse("2006-01-02", filter.StartDate)
		if err == nil {
			entryFilter.StartDate = &t
		}
	}
	if filter.EndDate != "" {
		t, err := time.Parse("2006-01-02", filter.EndDate)
		if err == nil {
			entryFilter.EndDate = &t
		}
	}

	entries, err := s.repo.FindEntriesByUserID(ctx, userID, entryFilter)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.CountByUserIDWithFilter(ctx, userID, entryFilter)
	if err != nil {
		return nil, err
	}

	result := make([]JournalEntryResponse, len(entries))
	for i, e := range entries {
		result[i] = toEntryResponse(&e)
	}

	return &ListResponse{
		Entries: result,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

func (s *service) GetTotalSummary(ctx context.Context, userID int64) (*JournalSummary, error) {
	totalDebit, err := s.repo.SumDebitByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	totalCredit, err := s.repo.SumCreditByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	count, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &JournalSummary{
		TotalDebit:       totalDebit,
		TotalCredit:      totalCredit,
		TransactionCount: count,
	}, nil
}

func (s *service) Create(ctx context.Context, userID int64, req CreateJournalRequest) (*JournalEntryResponse, error) {
	// Validate lines
	if len(req.Lines) < 2 {
		return nil, ErrInvalidLines
	}

	// Validate balanced
	totalDebit, totalCredit := 0.0, 0.0
	for _, l := range req.Lines {
		totalDebit += l.Debit
		totalCredit += l.Credit
	}
	if totalDebit != totalCredit {
		return nil, ErrNotBalanced
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format, expected YYYY-MM-DD")
	}

	// Validate all accounts are posting accounts
	for _, l := range req.Lines {
		acc, err := s.accountRepo.FindByID(ctx, l.AccountID)
		if err != nil {
			return nil, err
		}
		if acc == nil {
			return nil, errors.New("account not found")
		}
		if acc.UserID != userID {
			return nil, errors.New("account not found")
		}
		if !acc.CanPost() {
			return nil, ErrNotPostingAccount
		}
	}

	// Build lines
	lines := make([]JournalEntryLine, len(req.Lines))
	for i, l := range req.Lines {
		lines[i] = JournalEntryLine{
			AccountID: l.AccountID,
			Debit:     l.Debit,
			Credit:    l.Credit,
		}
	}

	entry := &JournalEntry{
		UserID:      userID,
		Date:        date,
		Description: req.Description,
		Lines:       lines,
	}

	err = s.repo.CreateEntry(ctx, entry)
	if err != nil {
		return nil, err
	}

	// Fetch the full entry with account info
	created, err := s.repo.FindEntryByID(ctx, entry.ID)
	if err != nil {
		return nil, err
	}

	resp := toEntryResponse(created)
	return &resp, nil
}

func (s *service) Update(ctx context.Context, userID int64, entryID int64, req UpdateJournalRequest) (*JournalEntryResponse, error) {
	existing, err := s.repo.FindEntryByID(ctx, entryID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if existing.UserID != userID {
		return nil, ErrNotFound
	}

	// Validate lines
	if len(req.Lines) < 2 {
		return nil, ErrInvalidLines
	}

	// Validate balanced
	totalDebit, totalCredit := 0.0, 0.0
	for _, l := range req.Lines {
		totalDebit += l.Debit
		totalCredit += l.Credit
	}
	if totalDebit != totalCredit {
		return nil, ErrNotBalanced
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format, expected YYYY-MM-DD")
	}

	// Validate all accounts are posting accounts
	for _, l := range req.Lines {
		acc, err := s.accountRepo.FindByID(ctx, l.AccountID)
		if err != nil {
			return nil, err
		}
		if acc == nil {
			return nil, errors.New("account not found")
		}
		if acc.UserID != userID {
			return nil, errors.New("account not found")
		}
		if !acc.CanPost() {
			return nil, ErrNotPostingAccount
		}
	}

	// Build lines
	lines := make([]JournalEntryLine, len(req.Lines))
	for i, l := range req.Lines {
		lines[i] = JournalEntryLine{
			AccountID: l.AccountID,
			Debit:     l.Debit,
			Credit:    l.Credit,
		}
	}

	existing.Date = date
	existing.Description = req.Description
	existing.Lines = lines

	err = s.repo.UpdateEntry(ctx, existing)
	if err != nil {
		return nil, err
	}

	// Fetch the full entry with account info
	updated, err := s.repo.FindEntryByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	resp := toEntryResponse(updated)
	return &resp, nil
}

func toEntryResponse(e *JournalEntry) JournalEntryResponse {
	lines := make([]JournalLineResponse, len(e.Lines))
	for i, l := range e.Lines {
		lines[i] = JournalLineResponse{
			ID:             l.ID,
			JournalEntryID: l.JournalEntryID,
			AccountID:      l.AccountID,
			Debit:          l.Debit,
			Credit:         l.Credit,
		}
	}

	return JournalEntryResponse{
		ID:          e.ID,
		Date:        e.Date.Format("2006-01-02"),
		Description: e.Description,
		Lines:       lines,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.Format(time.RFC3339),
	}
}
