package account

import (
	"testing"
)

func testAssetRow(id int64, code string, level int8, parentID *int64, balance float64) AssetBalanceRow {
	return AssetBalanceRow{
		Account: Account{
			ID:       id,
			UserID:   1,
			Code:     code,
			Name:     code,
			Type:     AccountTypeAsset,
			Level:    level,
			ParentID: parentID,
		},
		Balance: balance,
	}
}

func TestSummarizeAssetRows(t *testing.T) {
	rows := []AssetBalanceRow{
		testAssetRow(1, "1.00.00.00", 0, nil, 0),
		testAssetRow(2, "1.01.00.00", 1, ptrInt64(1), 0),
		testAssetRow(3, "1.01.01.00", 2, ptrInt64(2), 0),
		testAssetRow(4, "1.01.01.01", 3, ptrInt64(3), 100),
		testAssetRow(5, "1.01.01.02", 3, ptrInt64(3), 50),
		testAssetRow(6, "1.02.01.00", 2, ptrInt64(1), 0),
		testAssetRow(7, "1.02.01.01", 3, ptrInt64(6), 200),
	}

	summary := summarizeAssetRows(rows)
	if summary.TotalAsset != 350 {
		t.Fatalf("expected totalAsset 350, got %v", summary.TotalAsset)
	}
	if summary.CurrentAsset != 150 {
		t.Fatalf("expected currentAsset 150 (1.01.* only), got %v", summary.CurrentAsset)
	}
	if summary.AccountCount != 3 {
		t.Fatalf("expected accountCount 3, got %d", summary.AccountCount)
	}
}

func TestSummarizeAssetRowsEmpty(t *testing.T) {
	summary := summarizeAssetRows(nil)
	if summary.TotalAsset != 0 || summary.CurrentAsset != 0 || summary.AccountCount != 0 {
		t.Fatalf("expected zero summary, got %+v", summary)
	}
}

func TestBuildAssetGroups(t *testing.T) {
	rows := []AssetBalanceRow{
		testAssetRow(1, "1.00.00.00", 0, nil, 0),
		testAssetRow(2, "1.01.00.00", 1, ptrInt64(1), 0),
		testAssetRow(3, "1.01.01.00", 2, ptrInt64(2), 0),
		testAssetRow(4, "1.01.01.01", 3, ptrInt64(3), 100),
		testAssetRow(5, "1.01.01.02", 3, ptrInt64(3), 50),
		// Empty level-2 group (no posting children) must be skipped.
		testAssetRow(6, "1.01.02.00", 2, ptrInt64(2), 0),
		testAssetRow(7, "1.02.01.00", 2, ptrInt64(1), 0),
		testAssetRow(8, "1.02.01.01", 3, ptrInt64(7), 200),
		// Orphan posting (missing parent) must not crash nor appear.
		testAssetRow(9, "1.01.03.01", 3, ptrInt64(999), 999),
	}

	groups := buildAssetGroups(rows)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups (empty + orphan excluded), got %d", len(groups))
	}
	if groups[0].Code != "1.01.01.00" {
		t.Fatalf("expected first group 1.01.01.00, got %s", groups[0].Code)
	}
	if len(groups[0].Accounts) != 2 {
		t.Fatalf("expected 2 accounts in 1.01.01.00, got %d", len(groups[0].Accounts))
	}
	if groups[0].TotalBalance != 150 {
		t.Fatalf("expected totalBalance 150, got %v", groups[0].TotalBalance)
	}
	if groups[1].Code != "1.02.01.00" || groups[1].TotalBalance != 200 {
		t.Fatalf("unexpected second group: %+v", groups[1])
	}
}
