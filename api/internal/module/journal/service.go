package journal

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"pengbook/api/internal/database"
	"pengbook/api/internal/module/account"
	"pengbook/api/pkg/logger"
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
	// GetAllScrollView returns paginated journal entries with cursor-based pagination.
	GetAllScrollView(ctx context.Context, userID int64, filter ListRequest) (*CursorPageResponse, error)

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

func (s *service) GetAllScrollView(ctx context.Context, userID int64, filter ListRequest) (*CursorPageResponse, error) {
	log := logger.FromContext(ctx)

	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	entryFilter := EntryFilter{
		Limit:      limit,
		AccountIDs: filter.AccountIDs,
	}

	// Parse cursor
	if filter.Cursor != nil && *filter.Cursor != "" {
		parts := strings.SplitN(*filter.Cursor, "_", 2)
		if len(parts) == 2 {
			t, err := time.Parse("2006-01-02T15:04:05.000Z", parts[0])
			if err == nil {
				entryFilter.CursorDatetime = &t
			}else {
				log.Warn("journal GetAllScrollView: failed to parse cursor datetime", "cursor", *filter.Cursor, "error", err)
			}
			id, err := strconv.ParseInt(parts[1], 10, 64)
			if err == nil {
				entryFilter.CursorID = &id
			} else {
				log.Warn("journal GetAllScrollView: failed to parse cursor ID", "cursor", *filter.Cursor, "error", err)
			}
		}
	}

	// Parse date filters
	if filter.StartDate != "" {
		t, err := time.Parse(time.RFC3339, filter.StartDate)
		if err == nil {
			entryFilter.StartDate = &t
		}
	}
	if filter.EndDate != "" {
		t, err := time.Parse(time.RFC3339, filter.EndDate)
		if err == nil {
			entryFilter.EndDate = &t
		}
	}

	entries, err := s.repo.FindEntriesByUserIDWithNetEffect(ctx, userID, entryFilter)
	if err != nil {
		log.Error("journal GetAllScrollView: failed to fetch entries", "user_id", userID, "error", err)
		return nil, err
	}

	// Build next cursor
	var nextCursor *string
	if len(entries) == limit {
		last := entries[len(entries)-1]
		cursor := last.Datetime.UTC().Format("2006-01-02T15:04:05.000Z") + "_" + strconv.FormatInt(last.ID, 10)
		nextCursor = &cursor
	}

	return &CursorPageResponse{
		Data:       entries,
		NextCursor: nextCursor,
	}, nil
}

func (s *service) GetTotalSummary(ctx context.Context, userID int64) (*JournalSummary, error) {
	log := logger.FromContext(ctx)

	summary, err := s.repo.GetSummaryByUserID(ctx, userID)
	if err != nil {
		log.Error("journal GetTotalSummary: failed to get summary", "user_id", userID, "error", err)
		return nil, err
	}

	return summary, nil
}

func (s *service) Create(ctx context.Context, userID int64, req CreateJournalRequest) (*JournalEntryResponse, error) {
	log := logger.FromContext(ctx)

	// Validate lines
	if len(req.Lines) < 2 {
		log.Warn("journal create: invalid lines", "user_id", userID, "line_count", len(req.Lines))
		return nil, ErrInvalidLines
	}

	// Validate balanced
	totalDebit, totalCredit := 0.0, 0.0
	for _, l := range req.Lines {
		totalDebit += l.Debit
		totalCredit += l.Credit
	}
	if totalDebit != totalCredit {
		log.Warn("journal create: not balanced", "user_id", userID, "debit", totalDebit, "credit", totalCredit)
		return nil, ErrNotBalanced
	}

	// Parse date
	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		log.Warn("journal create: invalid date", "user_id", userID, "date", req.Date)
		return nil, errors.New("invalid date format, expected RFC3339 (e.g. 2026-04-21T10:30:00+07:00)")
	}

	// Validate all accounts are posting accounts (batch query)
	accountIDs := make([]int64, len(req.Lines))
	for i, l := range req.Lines {
		accountIDs[i] = l.AccountID
	}
	accounts, err := s.accountRepo.FindByIDs(ctx, accountIDs)
	if err != nil {
		log.Error("journal create: failed to find accounts", "user_id", userID, "account_ids", accountIDs, "error", err)
		return nil, err
	}

	for _, l := range req.Lines {
		acc, ok := accounts[l.AccountID]
		if !ok || acc == nil {
			log.Warn("journal create: account not found", "user_id", userID, "account_id", l.AccountID)
			return nil, errors.New("account not found")
		}
		if acc.UserID != userID {
			log.Warn("journal create: account not owned by user", "user_id", userID, "account_id", l.AccountID)
			return nil, errors.New("account not found")
		}
		if !acc.CanPost() {
			log.Warn("journal create: not a posting account", "user_id", userID, "account_id", l.AccountID, "level", acc.Level)
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

	// Use transaction to ensure atomicity of entry + lines creation
	var entryID int64
	err = s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		entry := &JournalEntry{
			UserID:      userID,
			Date:        date,
			Description: req.Description,
			Lines:       lines,
		}

		if err := s.repo.CreateEntry(ctx, entry); err != nil {
			return err
		}

		entryID = entry.ID
		return nil
	})

	if err != nil {
		log.Error("journal create: failed to create entry", "user_id", userID, "error", err)
		return nil, err
	}

	// Fetch the full entry with account info
	created, err := s.repo.FindEntryByID(ctx, entryID)
	if err != nil {
		log.Error("journal create: failed to fetch created entry", "user_id", userID, "entry_id", entryID, "error", err)
		return nil, err
	}

	log.Info("journal created", "user_id", userID, "entry_id", entryID, "line_count", len(req.Lines))
	resp := toEntryResponse(created)
	return &resp, nil
}

func (s *service) Update(ctx context.Context, userID int64, entryID int64, req UpdateJournalRequest) (*JournalEntryResponse, error) {
	log := logger.FromContext(ctx)

	existing, err := s.repo.FindEntryByID(ctx, entryID)
	if err != nil {
		log.Error("journal update: failed to find entry", "user_id", userID, "entry_id", entryID, "error", err)
		return nil, err
	}
	if existing == nil {
		log.Warn("journal update: entry not found", "user_id", userID, "entry_id", entryID)
		return nil, ErrNotFound
	}
	if existing.UserID != userID {
		log.Warn("journal update: entry not owned by user", "user_id", userID, "entry_id", entryID)
		return nil, ErrNotFound
	}

	// Validate lines
	if len(req.Lines) < 2 {
		log.Warn("journal update: invalid lines", "user_id", userID, "entry_id", entryID, "line_count", len(req.Lines))
		return nil, ErrInvalidLines
	}

	// Validate balanced
	totalDebit, totalCredit := 0.0, 0.0
	for _, l := range req.Lines {
		totalDebit += l.Debit
		totalCredit += l.Credit
	}
	if totalDebit != totalCredit {
		log.Warn("journal update: not balanced", "user_id", userID, "entry_id", entryID, "debit", totalDebit, "credit", totalCredit)
		return nil, ErrNotBalanced
	}

	// Parse date
	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		log.Warn("journal update: invalid date", "user_id", userID, "entry_id", entryID, "date", req.Date)
		return nil, errors.New("invalid date format, expected RFC3339 (e.g. 2026-04-21T10:30:00+07:00)")
	}

	// Validate all accounts are posting accounts (batch query)
	accountIDs := make([]int64, len(req.Lines))
	for i, l := range req.Lines {
		accountIDs[i] = l.AccountID
	}
	accounts, err := s.accountRepo.FindByIDs(ctx, accountIDs)
	if err != nil {
		log.Error("journal update: failed to find accounts", "user_id", userID, "account_ids", accountIDs, "error", err)
		return nil, err
	}

	for _, l := range req.Lines {
		acc, ok := accounts[l.AccountID]
		if !ok || acc == nil {
			log.Warn("journal update: account not found", "user_id", userID, "account_id", l.AccountID)
			return nil, errors.New("account not found")
		}
		if acc.UserID != userID {
			log.Warn("journal update: account not owned by user", "user_id", userID, "account_id", l.AccountID)
			return nil, errors.New("account not found")
		}
		if !acc.CanPost() {
			log.Warn("journal update: not a posting account", "user_id", userID, "account_id", l.AccountID, "level", acc.Level)
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

	// Use transaction to ensure atomicity of update operations
	err = s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		existing.Date = date
		existing.Description = req.Description
		existing.Lines = lines

		return s.repo.UpdateEntry(ctx, existing)
	})

	if err != nil {
		log.Error("journal update: failed to update entry", "user_id", userID, "entry_id", entryID, "error", err)
		return nil, err
	}

	// Fetch the full entry with account info
	updated, err := s.repo.FindEntryByID(ctx, entryID)
	if err != nil {
		log.Error("journal update: failed to fetch updated entry", "user_id", userID, "entry_id", entryID, "error", err)
		return nil, err
	}

	log.Info("journal updated", "user_id", userID, "entry_id", entryID, "line_count", len(req.Lines))
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
		Date:        e.Date.Format(time.RFC3339),
		Description: e.Description,
		Lines:       lines,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.Format(time.RFC3339),
	}
}
