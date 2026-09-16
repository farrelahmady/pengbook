package account

import (
	"context"
	"strings"

	"pengbook/api/pkg/logger"
)

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
