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

	// Delete removes a journal entry and reverses its balance effect.
	Delete(ctx context.Context, userID int64, entryID int64) error

	// CreateBulk creates multiple journal entries from a bulk upload.
	CreateBulk(ctx context.Context, userID int64, req CreateBulkJournalRequest) (int64, error)

	// GenerateTemplate generates an Excel template with user's posting accounts.
	GenerateTemplate(ctx context.Context, userID int64) ([]byte, error)
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

	// Delta update: only touch the accounts in this entry (O(lines)),
	// instead of re-summing their entire history.
	deltas := buildBalanceDeltas(lines, accounts)

	// Use transaction to ensure atomicity of entry + lines + balance creation
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

		// Balance is part of the write path now (not best-effort healing),
		// so a failure rolls the entry back instead of drifting silently.
		if err := s.repo.ApplyBalanceDelta(ctx, deltas); err != nil {
			log.Error("journal create: failed to apply balance delta", "user_id", userID, "error", err)
			return err
		}

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

	// Validate all accounts are posting accounts (batch query).
	// Lookup covers old + new accounts: new ones for validation, old ones for
	// sign factors when reversing their deltas.
	accountIDs := make([]int64, len(req.Lines))
	for i, l := range req.Lines {
		accountIDs[i] = l.AccountID
	}
	lookupSet := make(map[int64]bool, len(accountIDs)+len(existing.Lines))
	for _, id := range accountIDs {
		lookupSet[id] = true
	}
	for _, l := range existing.Lines {
		lookupSet[l.AccountID] = true
	}
	lookupIDs := make([]int64, 0, len(lookupSet))
	for id := range lookupSet {
		lookupIDs = append(lookupIDs, id)
	}
	accounts, err := s.accountRepo.FindByIDs(ctx, lookupIDs)
	if err != nil {
		log.Error("journal update: failed to find accounts", "user_id", userID, "account_ids", lookupIDs, "error", err)
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

	// Delta update: reverse the old lines, apply the new lines, merged per
	// account. Covers account changes (A -> B reverses A, applies B) and
	// amount-only changes with a single atomic UPSERT.
	oldDeltas := buildBalanceDeltas(existing.Lines, accounts)
	newDeltas := buildBalanceDeltas(lines, accounts)
	merged := make(map[int64]float64, len(oldDeltas)+len(newDeltas))
	for id, d := range oldDeltas {
		merged[id] -= d
	}
	for id, d := range newDeltas {
		merged[id] += d
	}

	// Use transaction to ensure atomicity of update operations
	err = s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		existing.Date = date
		existing.Description = req.Description
		existing.Lines = lines

		if err := s.repo.UpdateEntry(ctx, existing); err != nil {
			return err
		}

		// Balance is part of the write path now (not best-effort healing),
		// so a failure rolls the entry back instead of drifting silently.
		if err := s.repo.ApplyBalanceDelta(ctx, merged); err != nil {
			log.Error("journal update: failed to apply balance delta", "user_id", userID, "entry_id", entryID, "error", err)
			return err
		}

		return nil
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

func (s *service) Delete(ctx context.Context, userID int64, entryID int64) error {
	log := logger.FromContext(ctx)

	existing, err := s.repo.FindEntryByID(ctx, entryID)
	if err != nil {
		log.Error("journal delete: failed to find entry", "user_id", userID, "entry_id", entryID, "error", err)
		return err
	}
	if existing == nil {
		log.Warn("journal delete: entry not found", "user_id", userID, "entry_id", entryID)
		return ErrNotFound
	}
	if existing.UserID != userID {
		log.Warn("journal delete: entry not owned by user", "user_id", userID, "entry_id", entryID)
		return ErrNotFound
	}

	// Reverse the entry's balance effect: negate its signed deltas.
	lookupIDs := make([]int64, 0, len(existing.Lines))
	for _, l := range existing.Lines {
		lookupIDs = append(lookupIDs, l.AccountID)
	}
	accounts, err := s.accountRepo.FindByIDs(ctx, lookupIDs)
	if err != nil {
		log.Error("journal delete: failed to find accounts", "user_id", userID, "entry_id", entryID, "error", err)
		return err
	}
	reversal := buildBalanceDeltas(existing.Lines, accounts)
	for id, d := range reversal {
		reversal[id] = -d
	}

	err = s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.DeleteEntry(ctx, entryID); err != nil {
			return err
		}

		// Balance is part of the write path now (not best-effort healing),
		// so a failure rolls the delete back instead of drifting silently.
		if err := s.repo.ApplyBalanceDelta(ctx, reversal); err != nil {
			log.Error("journal delete: failed to apply balance delta", "user_id", userID, "entry_id", entryID, "error", err)
			return err
		}

		return nil
	})
	if err != nil {
		log.Error("journal delete: failed to delete entry", "user_id", userID, "entry_id", entryID, "error", err)
		return err
	}

	log.Info("journal deleted", "user_id", userID, "entry_id", entryID)
	return nil
}

func (s *service) CreateBulk(ctx context.Context, userID int64, req CreateBulkJournalRequest) (int64, error) {
	log := logger.FromContext(ctx)

	if len(req.Entries) == 0 {
		return 0, errors.New("no entries provided")
	}

	// Collect all account codes from all entries for batch lookup
	allAccountCodes := make(map[string]bool)
	for _, entry := range req.Entries {
		for _, line := range entry.Lines {
			if line.AccountCode != "" {
				allAccountCodes[line.AccountCode] = true
			}
		}
	}

	accountCodes := make([]string, 0, len(allAccountCodes))
	for code := range allAccountCodes {
		accountCodes = append(accountCodes, code)
	}

	// Find accounts by codes
	accountsByCode, err := s.accountRepo.FindByCodes(ctx, accountCodes)
	if err != nil {
		log.Error("journal create bulk: failed to find accounts by codes", "user_id", userID, "error", err)
		return 0, err
	}

	// Validate all accounts exist, are owned by user, and are posting accounts
	for code, acc := range accountsByCode {
		if acc == nil {
			log.Warn("journal create bulk: account not found by code", "user_id", userID, "account_code", code)
			return 0, errors.New("account not found: " + code)
		}
		if acc.UserID != userID {
			log.Warn("journal create bulk: account not owned by user", "user_id", userID, "account_code", code)
			return 0, errors.New("account not found: " + code)
		}
		if !acc.CanPost() {
			log.Warn("journal create bulk: not a posting account", "user_id", userID, "account_code", code, "level", acc.Level)
			return 0, ErrNotPostingAccount
		}
	}

	// Build all entries with resolved account IDs
	entries := make([]JournalEntry, 0, len(req.Entries))
	allAccountIDs := make(map[int64]bool)
	for _, entryReq := range req.Entries {
		// Validate lines
		if len(entryReq.Lines) < 2 {
			log.Warn("journal create bulk: invalid lines", "user_id", userID, "line_count", len(entryReq.Lines))
			return 0, ErrInvalidLines
		}

		// Validate balanced
		totalDebit, totalCredit := 0.0, 0.0
		for _, l := range entryReq.Lines {
			totalDebit += l.Debit
			totalCredit += l.Credit
		}
		if totalDebit != totalCredit {
			log.Warn("journal create bulk: not balanced", "user_id", userID, "debit", totalDebit, "credit", totalCredit)
			return 0, ErrNotBalanced
		}

		// Parse date
		date, err := time.Parse(time.RFC3339, entryReq.Date)
		if err != nil {
			log.Warn("journal create bulk: invalid date", "user_id", userID, "date", entryReq.Date)
			return 0, errors.New("invalid date format, expected RFC3339 (e.g. 2026-04-21T10:30:00+07:00)")
		}

		// Build lines with resolved account IDs
		lines := make([]JournalEntryLine, len(entryReq.Lines))
		for i, l := range entryReq.Lines {
			var accountID int64
			if l.AccountID > 0 {
				accountID = l.AccountID
			} else if l.AccountCode != "" {
				if acc, ok := accountsByCode[l.AccountCode]; ok && acc != nil {
					accountID = acc.ID
				} else {
					return 0, errors.New("account not found: " + l.AccountCode)
				}
			} else {
				return 0, errors.New("account id or account code is required")
			}
			allAccountIDs[accountID] = true
			lines[i] = JournalEntryLine{
				AccountID: accountID,
				Debit:     l.Debit,
				Credit:    l.Credit,
			}
		}

		entries = append(entries, JournalEntry{
			UserID:      userID,
			Date:        date,
			Description: entryReq.Description,
			Lines:       lines,
		})
	}

	// Resolve account types for every touched account. Codes are already in
	// accountsByCode; IDs passed directly are fetched so ownership and posting
	// level get the same validation as single create (previously unchecked).
	accountsByID := make(map[int64]*account.Account, len(allAccountIDs))
	for _, acc := range accountsByCode {
		if acc != nil {
			accountsByID[acc.ID] = acc
		}
	}
	missing := make([]int64, 0)
	for id := range allAccountIDs {
		if _, ok := accountsByID[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		fetched, err := s.accountRepo.FindByIDs(ctx, missing)
		if err != nil {
			log.Error("journal create bulk: failed to find accounts by ids", "user_id", userID, "error", err)
			return 0, err
		}
		for id, acc := range fetched {
			accountsByID[id] = acc
		}
	}
	for id := range allAccountIDs {
		acc := accountsByID[id]
		if acc == nil {
			log.Warn("journal create bulk: account not found by id", "user_id", userID, "account_id", id)
			return 0, errors.New("account not found")
		}
		if acc.UserID != userID {
			log.Warn("journal create bulk: account not owned by user", "user_id", userID, "account_id", id)
			return 0, errors.New("account not found")
		}
		if !acc.CanPost() {
			log.Warn("journal create bulk: not a posting account", "user_id", userID, "account_id", id, "level", acc.Level)
			return 0, ErrNotPostingAccount
		}
	}

	// Aggregate signed deltas across all entries: one atomic UPSERT instead of
	// re-summing each account's entire history.
	bulkDeltas := make(map[int64]float64, len(allAccountIDs))
	for _, e := range entries {
		for _, l := range e.Lines {
			acc := accountsByID[l.AccountID]
			if acc == nil {
				continue
			}
			bulkDeltas[l.AccountID] += (l.Debit - l.Credit) * balanceSignFactor(acc.Type)
		}
	}

	// Use transaction to ensure atomicity
	err = s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.CreateEntries(ctx, entries); err != nil {
			return err
		}

		// Balance is part of the write path now (not best-effort healing),
		// so a failure rolls the entries back instead of drifting silently.
		if err := s.repo.ApplyBalanceDelta(ctx, bulkDeltas); err != nil {
			log.Error("journal create bulk: failed to apply balance delta", "user_id", userID, "error", err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Error("journal create bulk: failed to create entries", "user_id", userID, "error", err)
		return 0, err
	}

	log.Info("journal entries created in bulk", "user_id", userID, "count", len(entries))
	return int64(len(entries)), nil
}

func (s *service) GenerateTemplate(ctx context.Context, userID int64) ([]byte, error) {
	log := logger.FromContext(ctx)

	// Fetch user's posting accounts
	accounts, err := s.accountRepo.FindPostingByUserID(ctx, userID)
	if err != nil {
		log.Error("journal GenerateTemplate: failed to fetch posting accounts", "user_id", userID, "error", err)
		return nil, err
	}

	// Convert to AccountInfo for template generation
	accountInfos := make([]AccountInfo, len(accounts))
	for i, acc := range accounts {
		accountInfos[i] = AccountInfo{
			Code:  acc.Code,
			Name:  acc.Name,
			Type:  string(acc.Type),
			Level: acc.Level,
		}
	}

	// Generate Excel template
	template, err := GenerateTemplateWithAccounts(accountInfos)
	if err != nil {
		log.Error("journal GenerateTemplate: failed to generate template", "user_id", userID, "error", err)
		return nil, err
	}

	log.Info("journal template generated", "user_id", userID, "account_count", len(accounts))
	return template, nil
}

// balanceSignFactor mirrors recalculate_account_balance in SQL: balance is stored
// as (debit - credit) * sign, where revenue/liability/equity carry a negative
// sign so a normal credit balance reads positive.
func balanceSignFactor(t account.AccountType) float64 {
	switch t {
	case account.AccountTypeLiability, account.AccountTypeEquity, account.AccountTypeRevenue:
		return -1
	default: // ASSET, EXPENSE, OTHER
		return 1
	}
}

// buildBalanceDeltas aggregates signed deltas per account for the given lines.
// Signed delta per line = (debit - credit) * signFactor(account type).
// Lines whose account info is missing are skipped defensively (callers validate
// beforehand, so this only guards against stale references).
func buildBalanceDeltas(lines []JournalEntryLine, accounts map[int64]*account.Account) map[int64]float64 {
	deltas := make(map[int64]float64, len(lines))
	for _, l := range lines {
		acc := accounts[l.AccountID]
		if acc == nil {
			continue
		}
		deltas[l.AccountID] += (l.Debit - l.Credit) * balanceSignFactor(acc.Type)
	}
	return deltas
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
