package account

import (
	"testing"
	"time"
)

func ptrInt64(v int64) *int64 { return &v }

func testAccount(id int64, code string, typ AccountType, level int8, parentID *int64) Account {
	return Account{
		ID:        id,
		UserID:    1,
		Code:      code,
		Name:      code,
		Type:      typ,
		Level:     level,
		ParentID:  parentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Regression test for the value-copy bug: with ORDER BY code (parents
// first), roots must still contain the full 4-level subtree.
func TestBuildTreePreservesHierarchy(t *testing.T) {
	accounts := []Account{
		testAccount(1, "1.00.00.00", AccountTypeAsset, 0, nil),
		testAccount(2, "1.01.00.00", AccountTypeAsset, 1, ptrInt64(1)),
		testAccount(3, "1.01.01.00", AccountTypeAsset, 2, ptrInt64(2)),
		testAccount(4, "1.01.01.01", AccountTypeAsset, 3, ptrInt64(3)),
	}

	roots := buildTree(accounts)
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	l1 := roots[0].Children
	if len(l1) != 1 {
		t.Fatalf("expected 1 level-1 child, got %d", len(l1))
	}
	l2 := l1[0].Children
	if len(l2) != 1 {
		t.Fatalf("expected 1 level-2 child, got %d", len(l2))
	}
	l3 := l2[0].Children
	if len(l3) != 1 || l3[0].Code != "1.01.01.01" || !l3[0].IsPosting {
		t.Fatalf("expected posting leaf 1.01.01.01, got %+v", l3)
	}
}

// Tree must not depend on input order (children first must also work)
// and orphans must surface as roots instead of being dropped.
func TestBuildTreeOrderIndependentAndOrphan(t *testing.T) {
	accounts := []Account{
		testAccount(4, "1.01.01.01", AccountTypeAsset, 3, ptrInt64(3)),
		testAccount(3, "1.01.01.00", AccountTypeAsset, 2, ptrInt64(2)),
		testAccount(2, "1.01.00.00", AccountTypeAsset, 1, ptrInt64(1)),
		testAccount(1, "1.00.00.00", AccountTypeAsset, 0, nil),
		testAccount(9, "9.00.00.00", AccountTypeOther, 0, ptrInt64(999)),
	}

	roots := buildTree(accounts)
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots (1 tree + 1 orphan), got %d", len(roots))
	}
	if len(roots[0].Children[0].Children[0].Children) != 1 {
		t.Fatalf("reverse-order tree lost children: %+v", roots[0])
	}
}

// groupByType must count the whole subtree per type group.
func TestGroupByTypeCountsSubtree(t *testing.T) {
	accounts := []Account{
		testAccount(1, "1.00.00.00", AccountTypeAsset, 0, nil),
		testAccount(2, "1.01.00.00", AccountTypeAsset, 1, ptrInt64(1)),
		testAccount(3, "1.01.01.00", AccountTypeAsset, 2, ptrInt64(2)),
		testAccount(4, "1.01.01.01", AccountTypeAsset, 3, ptrInt64(3)),
	}

	groups := groupByType(buildTree(accounts))
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Count != 4 {
		t.Fatalf("expected count 4, got %d", groups[0].Count)
	}
}
