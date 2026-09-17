package account

import (
	"context"
	"strings"
	"time"

	"pengbook/api/pkg/logger"
)

// JournalMutator is the minimal interface for creating journal entries.
// Defined here to avoid importing the journal package (which would create
// an import cycle: account -> journal -> account).
// The concrete implementation is postgres.journalRepository.
type JournalMutator interface {
	CreateEntry(ctx context.Context, userID int64, date time.Time, description string, lines []JournalEntryLineInput) (int64, error)
	ApplyBalanceDelta(ctx context.Context, deltas map[int64]float64) error
	InsertAuditLog(ctx context.Context, userID int64, journalEntryID *int64, action string, newValues JSONB) error
}

// JournalEntryLineInput is a basic input struct for creating journal lines.
// Used by journalMutator to avoid importing journal package.
type JournalEntryLineInput struct {
	AccountID int64
	Debit     float64
	Credit    float64
}

// currentAssetPrefix marks the Current Assets subtree (1.01.*).
// Derived from the code prefix (not parent traversal) so it stays
// consistent with the generated type/level columns.
const currentAssetPrefix = "1.01."

// GetAssetSummary returns the lightweight Asset aggregates for a user.
// Only posting (level 3) balances are summed; header rows carry
// COALESCE(balance, 0) and contribute nothing.
func (s *service) GetAssetSummary(ctx context.Context, userID int64) (*AssetSummary, error) {
	log := logger.FromContext(ctx)

	rows, err := s.repo.FindAssetWithBalances(ctx, userID)
	if err != nil {
		log.Error("asset summary: failed to find assets", "user_id", userID, "error", err)
		return nil, err
	}

	summary := summarizeAssetRows(rows)
	log.Info("asset summary fetched", "user_id", userID,
		"total_asset", summary.TotalAsset, "current_asset", summary.CurrentAsset,
		"account_count", summary.AccountCount)
	return &summary, nil
}

// GetAssetGroups returns asset posting accounts grouped under their
// level-2 ancestor (e.g. 1.01.01 Bank), with per-group totals.
// Groups without posting accounts are skipped.
func (s *service) GetAssetGroups(ctx context.Context, userID int64) (*AssetGroups, error) {
	log := logger.FromContext(ctx)

	rows, err := s.repo.FindAssetWithBalances(ctx, userID)
	if err != nil {
		log.Error("asset groups: failed to find assets", "user_id", userID, "error", err)
		return nil, err
	}

	groups := buildAssetGroups(rows)
	log.Info("asset groups fetched", "user_id", userID, "groups", len(groups))
	return &AssetGroups{Groups: groups}, nil
}

// summarizeAssetRows aggregates posting balances from asset rows.
// Pure function (no I/O) so it is unit-testable.
func summarizeAssetRows(rows []AssetBalanceRow) AssetSummary {
	var summary AssetSummary
	for _, r := range rows {
		if r.Account.Level != 3 {
			continue
		}
		summary.TotalAsset += r.Balance
		if strings.HasPrefix(r.Account.Code, currentAssetPrefix) {
			summary.CurrentAsset += r.Balance
		}
		summary.AccountCount++
	}
	return summary
}

// buildAssetGroups collects posting accounts under their level-2 ancestor.
// Rows must be ordered by code (parents before children); groups keep
// that order. Pure function (no I/O) so it is unit-testable.
func buildAssetGroups(rows []AssetBalanceRow) []AssetGroup {
	byID := make(map[int64]*AssetBalanceRow, len(rows))
	for i := range rows {
		byID[rows[i].Account.ID] = &rows[i]
	}

	// Level-2 ancestors become groups, in code order.
	groupIdx := make(map[int64]int)
	var groups []AssetGroup
	for i := range rows {
		a := rows[i].Account
		if a.Level != 2 {
			continue
		}
		groupIdx[a.ID] = len(groups)
		groups = append(groups, AssetGroup{
			ID:       a.ID,
			Code:     a.Code,
			Name:     a.Name,
			Accounts: []AssetAccount{},
		})
	}

	// Attach each posting account to its level-2 ancestor.
	for i := range rows {
		a := rows[i].Account
		if a.Level != 3 {
			continue
		}
		ancestor := findLevel2Ancestor(byID, &rows[i])
		if ancestor == nil {
			continue
		}
		idx, ok := groupIdx[ancestor.Account.ID]
		if !ok {
			continue
		}
		groups[idx].Accounts = append(groups[idx].Accounts, AssetAccount{
			ID:        a.ID,
			Code:      a.Code,
			Name:      a.Name,
			Balance:   rows[i].Balance,
			IsPosting: true,
		})
		groups[idx].TotalBalance += rows[i].Balance
	}

	// Skip groups without posting accounts.
	result := groups[:0]
	for _, g := range groups {
		if len(g.Accounts) > 0 {
			result = append(result, g)
		}
	}
	return result
}

// findLevel2Ancestor walks parents up to the level-2 ancestor.
// Returns nil for orphans or rows already at/above level 2.
func findLevel2Ancestor(byID map[int64]*AssetBalanceRow, row *AssetBalanceRow) *AssetBalanceRow {
	cur := row
	for cur != nil && cur.Account.Level > 2 {
		if cur.Account.ParentID == nil {
			return nil
		}
		parent, ok := byID[*cur.Account.ParentID]
		if !ok {
			return nil
		}
		cur = parent
	}
	if cur == nil || cur.Account.Level != 2 {
		return nil
	}
	return cur
}

// CreateAsset creates a new asset posting account under an ASSET level-2 parent.
// Balance starts at 0; use AdjustAssetBalance to set the opening balance.
func (s *service) CreateAsset(ctx context.Context, userID int64, req CreateAssetRequest) (*AccountResponse, error) {
	log := logger.FromContext(ctx)

	parent, err := s.repo.FindByID(ctx, req.ParentID)
	if err != nil {
		log.Error("create asset: failed to find parent", "user_id", userID, "parent_id", req.ParentID, "error", err)
		return nil, err
	}
	if parent == nil || parent.UserID != userID {
		log.Warn("create asset: parent not found", "user_id", userID, "parent_id", req.ParentID)
		return nil, ErrParentNotFound
	}
	if parent.Type != AccountTypeAsset {
		log.Warn("create asset: parent is not ASSET type", "user_id", userID, "parent_id", req.ParentID, "type", parent.Type)
		return nil, ErrParentNotAsset
	}
	if parent.Level != 2 {
		log.Warn("create asset: parent is not level 2", "user_id", userID, "parent_id", req.ParentID, "level", parent.Level)
		return nil, ErrParentNotLevel2
	}

	code, err := s.nextChildCode(ctx, userID, parent)
	if err != nil {
		log.Error("create asset: failed to generate code", "user_id", userID, "parent_id", req.ParentID, "error", err)
		return nil, err
	}

	a := &Account{
		UserID:   userID,
		Code:     code,
		Name:     req.Name,
		ParentID: &req.ParentID,
	}

	var created *Account
	err = s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.Create(ctx, a); err != nil {
			return err
		}
		created = a

		// Asset posting accounts track a cached balance (starts at 0).
		if err := s.repo.EnsureBalance(ctx, a.ID); err != nil {
			return err
		}

		// Audit log
		return s.repo.InsertAuditLog(ctx, &AccountAuditLog{
			UserID:    userID,
			AccountID: &a.ID,
			Action:    string(AccountActionCreated),
			NewValues: &JSONB{"code": a.Code, "name": a.Name, "type": a.Type, "level": a.Level},
		})
	})
	if err != nil {
		log.Error("create asset: failed to create asset", "user_id", userID, "parent_id", req.ParentID, "error", err)
		return nil, err
	}

	log.Info("asset created", "user_id", userID, "asset_id", a.ID, "code", a.Code, "name", a.Name)
	return toResponse(created), nil
}

// AdjustAssetBalance adjusts the balance of an asset posting account to the
// desired balance via a balanced journal entry against the reclass adjustment
// account (6.01.01.01). The delta is calculated server-side:
// - desiredBalance > currentBalance: Dr Asset / Cr Reclass (amount = delta)
// - desiredBalance < currentBalance: Dr Reclass / Cr Asset (amount = delta)
// - desiredBalance == currentBalance: no journal entry, return current balance
func (s *service) AdjustAssetBalance(ctx context.Context, userID int64, assetID int64, req AdjustBalanceRequest) (*AdjustBalanceResponse, error) {
	log := logger.FromContext(ctx)

	// Validate asset exists and is an ASSET posting account
	asset, err := s.repo.FindByID(ctx, assetID)
	if err != nil {
		log.Error("adjust balance: failed to find asset", "user_id", userID, "asset_id", assetID, "error", err)
		return nil, err
	}
	if asset == nil || asset.UserID != userID {
		log.Warn("adjust balance: asset not found", "user_id", userID, "asset_id", assetID)
		return nil, ErrNotFound
	}
	if asset.Type != AccountTypeAsset || asset.Level != 3 {
		log.Warn("adjust balance: asset is not a posting account", "user_id", userID, "asset_id", assetID, "type", asset.Type, "level", asset.Level)
		return nil, ErrAssetNotPosting
	}

	// Get current balance
	currentBalance := float64(0)
	rows, err := s.repo.FindAssetWithBalances(ctx, userID)
	if err != nil {
		log.Error("adjust balance: failed to fetch balances", "user_id", userID, "error", err)
		return nil, err
	}
	for _, row := range rows {
		if row.Account.ID == assetID {
			currentBalance = row.Balance
			break
		}
	}

	// If desired balance equals current, nothing to do
	if req.Balance == currentBalance {
		log.Info("adjust balance: no change needed", "user_id", userID, "asset_id", assetID, "balance", currentBalance)
		return &AdjustBalanceResponse{
			JournalEntryID: 0,
			Balance:       currentBalance,
		}, nil
	}

	// Calculate delta and direction
	delta := req.Balance - currentBalance
	log.Debug("adjust balance: calculated delta", "user_id", userID, "asset_id", assetID,
		"current", currentBalance, "desired", req.Balance, "delta", delta)

	// Ensure reclass chain exists: 6.00 -> 6.01 -> 6.01.01 -> 6.01.01.01
	reclassID, err := s.ensureReclassChain(ctx, userID)
	if err != nil {
		log.Error("adjust balance: failed to ensure reclass chain", "user_id", userID, "error", err)
		return nil, err
	}

	// Parse date or use now
	date := time.Now()
	if req.Date != "" {
		date, err = time.Parse(time.RFC3339, req.Date)
		if err != nil {
			log.Warn("adjust balance: invalid date, using now", "user_id", userID, "date", req.Date)
			date = time.Now()
		}
	}

	// Build description
	description := "Asset balance adjustment"
	if req.Note != "" {
		description = req.Note
	}

	// Build journal lines based on delta direction
	// delta > 0: asset needs to increase → Dr Asset / Cr Reclass
	// delta < 0: asset needs to decrease → Dr Reclass / Cr Asset
	var lines []JournalEntryLineInput
	if delta > 0 {
		// Dr Asset / Cr Reclass
		lines = []JournalEntryLineInput{
			{AccountID: assetID, Debit: delta, Credit: 0},
			{AccountID: reclassID, Debit: 0, Credit: delta},
		}
	} else {
		// Dr Reclass / Cr Asset (amount is positive)
		amount := -delta
		lines = []JournalEntryLineInput{
			{AccountID: reclassID, Debit: amount, Credit: 0},
			{AccountID: assetID, Debit: 0, Credit: amount},
		}
	}

	var entryID int64
	err = s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		// Create journal entry and lines
		entryID, err = s.journalRepo.CreateEntry(ctx, userID, date, description, lines)
		if err != nil {
			return err
		}

		// Build and apply balance deltas
		deltas := make(map[int64]float64)
		for _, line := range lines {
			deltas[line.AccountID] += line.Debit - line.Credit
		}
		if err := s.journalRepo.ApplyBalanceDelta(ctx, deltas); err != nil {
			log.Error("adjust balance: failed to apply delta", "user_id", userID, "error", err)
			return err
		}

		// Audit log for journal entry
		linesSnap := make([]map[string]interface{}, len(lines))
		for i, l := range lines {
			linesSnap[i] = map[string]interface{}{
				"account_id": l.AccountID,
				"debit":      l.Debit,
				"credit":     l.Credit,
			}
		}
		newValues := JSONB{
			"date":         date.Format(time.RFC3339),
			"description":  description,
			"asset_id":     assetID,
			"desired":      req.Balance,
			"current":      currentBalance,
			"delta":        delta,
			"reclass_id":   reclassID,
			"lines":        linesSnap,
		}
		return s.journalRepo.InsertAuditLog(ctx, userID, &entryID, "journal.created", newValues)
	})
	if err != nil {
		log.Error("adjust balance: failed", "user_id", userID, "asset_id", assetID, "error", err)
		return nil, err
	}

	log.Info("adjust balance: completed", "user_id", userID, "asset_id", assetID, "journal_entry_id", entryID,
		"from", currentBalance, "to", req.Balance, "delta", delta)
	return &AdjustBalanceResponse{
		JournalEntryID: entryID,
		Balance:       req.Balance,
	}, nil
}

// ensureReclassChain ensures the reclass adjustment chain exists for a user:
// 6.00.00.00 -> 6.01.00.00 -> 6.01.01.00 -> 6.01.01.01
// Uses INSERT ... ON CONFLICT DO NOTHING so it is idempotent.
// Returns the account ID of 6.01.01.01.
func (s *service) ensureReclassChain(ctx context.Context, userID int64) (int64, error) {
	log := logger.FromContext(ctx)

	// First ensure the roots exist (this also seeds 6.00 if not exists)
	if err := s.repo.SeedRoots(ctx, userID); err != nil {
		log.Error("ensure reclass chain: failed to seed roots", "user_id", userID, "error", err)
		return 0, err
	}

	// Check if 6.01.01.01 already exists
	accounts, err := s.repo.FindByCodes(ctx, []string{
		"6.01.01.01",
	})
	if err != nil {
		log.Error("ensure reclass chain: failed to find reclass account", "user_id", userID, "error", err)
		return 0, err
	}
	if acc, ok := accounts["6.01.01.01"]; ok && acc.UserID == userID {
		return acc.ID, nil
	}

	// We need to create the chain 6.01 -> 6.01.01 -> 6.01.01.01
	// First find or create 6.01
	id6, err := s.ensureAccount(ctx, userID, "6.01.00.00", "Reclass", nil)
	if err != nil {
		log.Error("ensure reclass chain: failed to ensure 6.01", "user_id", userID, "error", err)
		return 0, err
	}

	// Then find or create 6.01.01
	id6101, err := s.ensureAccount(ctx, userID, "6.01.01.00", "Reclass Adjustments", &id6)
	if err != nil {
		log.Error("ensure reclass chain: failed to ensure 6.01.01", "user_id", userID, "error", err)
		return 0, err
	}

	// Finally create 6.01.01.01
	id, err := s.ensureAccount(ctx, userID, "6.01.01.01", "Reclass Adjustments", &id6101)
	if err != nil {
		log.Error("ensure reclass chain: failed to ensure 6.01.01.01", "user_id", userID, "error", err)
		return 0, err
	}

	log.Info("ensure reclass chain: created/found reclass chain", "user_id", userID, "reclass_id", id)
	return id, nil
}

// ensureAccount creates an account if it doesn't exist (by code, per user).
// Returns the account ID. Uses INSERT ... ON CONFLICT DO NOTHING.
func (s *service) ensureAccount(ctx context.Context, userID int64, code, name string, parentID *int64) (int64, error) {
	log := logger.FromContext(ctx)

	// Check if exists
	accounts, err := s.repo.FindByCodes(ctx, []string{code})
	if err != nil {
		return 0, err
	}
	if acc, ok := accounts[code]; ok && acc.UserID == userID {
		return acc.ID, nil
	}

	// Create it
	a := &Account{
		UserID:   userID,
		Code:     code,
		Name:     name,
		ParentID: parentID,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		log.Error("ensure account: failed to create", "user_id", userID, "code", code, "error", err)
		return 0, err
	}
	log.Debug("ensure account: created", "user_id", userID, "code", code, "account_id", a.ID)
	return a.ID, nil
}
