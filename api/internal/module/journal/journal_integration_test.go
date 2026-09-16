// Package journal_test contains integration tests for the journal module.
// These tests require an active PostgreSQL database (config read from .env).
package journal_test

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengbook/api/internal/database"
	"pengbook/api/internal/infrastructure/postgres"
	"pengbook/api/internal/module/account"
	"pengbook/api/internal/module/journal"
	"pengbook/api/internal/module/user"
)

// testContext holds all dependencies needed for journal tests.
type testContext struct {
	pool      *pgxpool.Pool
	userRepo  user.Repository
	accRepo   account.Repository
	jnlRepo   journal.Repository
	userSvc   user.Service
	accSvc    account.Service
	jnlSvc    journal.Service
}

// setup builds all dependencies for journal integration tests.
func setup(t *testing.T) *testContext {
	t.Helper()

	pool := database.NewTestPool(t)

	userRepo := postgres.NewUserRepository(pool)
	accRepo := postgres.NewAccountRepository(pool)
	jnlRepo := postgres.NewJournalRepository(pool)

	txMgr := postgres.NewTxManager(pool)
	userSvc := user.NewService(userRepo, txMgr)
	accSvc := account.NewService(accRepo, postgres.NewJournalMutatorAdapter(jnlRepo), txMgr)
	jnlSvc := journal.NewService(jnlRepo, accRepo, txMgr)

	return &testContext{
		pool:     pool,
		userRepo: userRepo,
		accRepo:  accRepo,
		jnlRepo:  jnlRepo,
		userSvc:  userSvc,
		accSvc:   accSvc,
		jnlSvc:   jnlSvc,
	}
}

// createUser creates a test user and returns the ID.
func createUser(t *testing.T, tc *testContext) int64 {
	t.Helper()
	ctx := context.Background()

	// Use random number to ensure uniqueness across parallel test runs
	uniqueID := fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(10000))
	u, err := tc.userSvc.Create(ctx, user.CreateUserRequest{
		Name:     "Test User " + uniqueID,
		Username: "tuser" + uniqueID,
		Email:    "test-" + uniqueID + "@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	t.Cleanup(func() {
		tc.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", u.ID)
	})

	return u.ID
}

// createAccount creates a test account with the exact code (building the
// parent chain first) and returns the ID. Fixtures bypass the service
// codegen and insert explicit codes via the repository.
func createAccount(t *testing.T, tc *testContext, userID int64, code string) int64 {
	t.Helper()
	ctx := context.Background()

	segs := strings.Split(code, ".")
	if len(segs) != 4 {
		t.Fatalf("invalid test account code: %s", code)
	}
	prefixes := []string{
		fmt.Sprintf("%s.00.00.00", segs[0]),
		fmt.Sprintf("%s.%s.00.00", segs[0], segs[1]),
		fmt.Sprintf("%s.%s.%s.00", segs[0], segs[1], segs[2]),
		code,
	}
	level := 0
	for _, s := range segs[1:] {
		if s != "00" {
			level++
		}
	}

	existing, err := tc.accRepo.FindByCodes(ctx, prefixes[:level+1])
	if err != nil {
		t.Fatalf("find accounts: %v", err)
	}

	var parentID *int64
	var lastID int64
	for _, p := range prefixes[:level+1] {
		if acc, ok := existing[p]; ok {
			id := acc.ID
			parentID = &id
			lastID = id
			continue
		}
		a := &account.Account{
			UserID:   userID,
			Code:     p,
			Name:     fmt.Sprintf("Account %s", p),
			ParentID: parentID,
		}
		if err := tc.accRepo.Create(ctx, a); err != nil {
			t.Fatalf("create account %s: %v", p, err)
		}
		id := a.ID
		parentID = &id
		lastID = id
	}

	return lastID
}

// ── Repository Tests ─────────────────────────────────────────────────────────

func TestRepository_CreateEntry(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	entry := &journal.JournalEntry{
		UserID:      userID,
		Date:        time.Now(),
		Description: "Test entry",
		Lines: []journal.JournalEntryLine{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}

	err := tc.jnlRepo.CreateEntry(ctx, entry)
	if err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}

	if entry.ID == 0 {
		t.Fatal("expected non-zero entry ID")
	}

	// Verify entry can be found
	found, err := tc.jnlRepo.FindEntryByID(ctx, entry.ID)
	if err != nil {
		t.Fatalf("FindEntryByID: %v", err)
	}
	if found == nil {
		t.Fatal("expected entry to be found")
	}
	if found.Description != "Test entry" {
		t.Errorf("expected description 'Test entry', got %s", found.Description)
	}
	if len(found.Lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(found.Lines))
	}

	// Cleanup
	tc.jnlRepo.DeleteEntry(ctx, entry.ID)
}

func TestRepository_FindEntryByID_NotFound(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()

	found, err := tc.jnlRepo.FindEntryByID(ctx, 999999)
	if err != nil {
		t.Fatalf("FindEntryByID: %v", err)
	}
	if found != nil {
		t.Fatal("expected nil for non-existent entry")
	}
}

func TestRepository_UpdateEntry(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Create entry
	entry := &journal.JournalEntry{
		UserID:      userID,
		Date:        time.Now(),
		Description: "Original description",
		Lines: []journal.JournalEntryLine{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}
	if err := tc.jnlRepo.CreateEntry(ctx, entry); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}

	// Update entry
	entry.Description = "Updated description"
	entry.Lines = []journal.JournalEntryLine{
		{AccountID: accID1, Debit: 200000, Credit: 0},
		{AccountID: accID2, Debit: 0, Credit: 200000},
	}

	if err := tc.jnlRepo.UpdateEntry(ctx, entry); err != nil {
		t.Fatalf("UpdateEntry: %v", err)
	}

	// Verify update
	found, err := tc.jnlRepo.FindEntryByID(ctx, entry.ID)
	if err != nil {
		t.Fatalf("FindEntryByID: %v", err)
	}
	if found.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got %s", found.Description)
	}
	if len(found.Lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(found.Lines))
	}

	// Cleanup
	tc.jnlRepo.DeleteEntry(ctx, entry.ID)
}

func TestRepository_DeleteEntry(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Create entry
	entry := &journal.JournalEntry{
		UserID:      userID,
		Date:        time.Now(),
		Description: "To be deleted",
		Lines: []journal.JournalEntryLine{
			{AccountID: accID1, Debit: 50000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 50000},
		},
	}
	if err := tc.jnlRepo.CreateEntry(ctx, entry); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}

	// Delete entry
	if err := tc.jnlRepo.DeleteEntry(ctx, entry.ID); err != nil {
		t.Fatalf("DeleteEntry: %v", err)
	}

	// Verify deletion
	found, err := tc.jnlRepo.FindEntryByID(ctx, entry.ID)
	if err != nil {
		t.Fatalf("FindEntryByID: %v", err)
	}
	if found != nil {
		t.Fatal("expected entry to be deleted")
	}
}

func TestRepository_GetMonthlySummaryByUserID(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accKas := createAccount(t, tc, userID, "1.01.01.01")
	accPendapatan := createAccount(t, tc, userID, "4.01.01.01")
	accBeban := createAccount(t, tc, userID, "5.01.01.01")

	// Two entries inside March 2026, one outside.
	entries := []journal.JournalEntry{
		{
			UserID:      userID,
			Date:        time.Date(2026, 3, 10, 12, 0, 0, 0, time.FixedZone("WIB", 7*3600)),
			Description: "Penjualan",
			Lines: []journal.JournalEntryLine{
				{AccountID: accKas, Debit: 500000, Credit: 0},
				{AccountID: accPendapatan, Debit: 0, Credit: 500000},
			},
		},
		{
			UserID:      userID,
			Date:        time.Date(2026, 3, 15, 12, 0, 0, 0, time.FixedZone("WIB", 7*3600)),
			Description: "Beban",
			Lines: []journal.JournalEntryLine{
				{AccountID: accBeban, Debit: 200000, Credit: 0},
				{AccountID: accKas, Debit: 0, Credit: 200000},
			},
		},
		{
			UserID:      userID,
			Date:        time.Date(2026, 4, 5, 12, 0, 0, 0, time.FixedZone("WIB", 7*3600)),
			Description: "Penjualan April",
			Lines: []journal.JournalEntryLine{
				{AccountID: accKas, Debit: 999000, Credit: 0},
				{AccountID: accPendapatan, Debit: 0, Credit: 999000},
			},
		},
	}
	for i := range entries {
		if err := tc.jnlRepo.CreateEntry(ctx, &entries[i]); err != nil {
			t.Fatalf("CreateEntry: %v", err)
		}
		defer tc.jnlRepo.DeleteEntry(ctx, entries[i].ID)
	}

	loc := time.FixedZone("WIB", 7*3600)
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, 0)

	summary, err := tc.jnlRepo.GetMonthlySummaryByUserID(ctx, userID, start, end)
	if err != nil {
		t.Fatalf("GetMonthlySummaryByUserID: %v", err)
	}

	if summary.Income != 500000 {
		t.Errorf("expected income 500000, got %v", summary.Income)
	}
	if summary.Expense != 200000 {
		t.Errorf("expected expense 200000, got %v", summary.Expense)
	}
	if summary.TransactionCount != 2 {
		t.Errorf("expected transactionCount 2, got %d", summary.TransactionCount)
	}
}

func TestRepository_FindByCodes(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Find by codes
	codes := []string{"1.01.01.01", "4.01.01.01"}
	accounts, err := tc.accRepo.FindByCodes(ctx, codes)
	if err != nil {
		t.Fatalf("FindByCodes: %v", err)
	}

	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts))
	}

	if acc, ok := accounts["1.01.01.01"]; !ok || acc.ID != accID1 {
		t.Errorf("expected account 1.01.01.01 with ID %d", accID1)
	}
	if acc, ok := accounts["4.01.01.01"]; !ok || acc.ID != accID2 {
		t.Errorf("expected account 4.01.01.01 with ID %d", accID2)
	}
}

func TestRepository_RecalculateAccountBalance(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Create entry
	entry := &journal.JournalEntry{
		UserID:      userID,
		Date:        time.Now(),
		Description: "Balance test",
		Lines: []journal.JournalEntryLine{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}
	if err := tc.jnlRepo.CreateEntry(ctx, entry); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}

	// Recalculate balances
	err := tc.jnlRepo.RecalculateAccountBalance(ctx, []int64{accID1, accID2}, userID)
	if err != nil {
		t.Fatalf("RecalculateAccountBalance: %v", err)
	}

	// Cleanup
	tc.jnlRepo.DeleteEntry(ctx, entry.ID)
}

// ── Service Tests ────────────────────────────────────────────────────────────

func TestService_Create_Success(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	req := journal.CreateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Service create test",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}

	created, err := tc.jnlSvc.Create(ctx, userID, req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if created.ID == 0 {
		t.Fatal("expected non-zero entry ID")
	}
	if created.Description != "Service create test" {
		t.Errorf("expected description 'Service create test', got %s", created.Description)
	}

	// Cleanup
	tc.jnlRepo.DeleteEntry(ctx, created.ID)
}

func TestService_Create_ErrNotBalanced(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	req := journal.CreateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Unbalanced entry",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 200000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}

	_, err := tc.jnlSvc.Create(ctx, userID, req)
	if err != journal.ErrNotBalanced {
		t.Fatalf("expected ErrNotBalanced, got %v", err)
	}
}

func TestService_Create_ErrInvalidLines(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")

	req := journal.CreateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Too few lines",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 100000, Credit: 0},
		},
	}

	_, err := tc.jnlSvc.Create(ctx, userID, req)
	if err != journal.ErrInvalidLines {
		t.Fatalf("expected ErrInvalidLines, got %v", err)
	}
}

func TestService_Create_ErrNotPostingAccount(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)

	// Create a header account (level < 3) - use valid 4-segment code
	headerID := createAccount(t, tc, userID, "1.01.00.00")
	headerAcc, err := tc.accRepo.FindByID(ctx, headerID)
	if err != nil || headerAcc == nil {
		t.Fatalf("find header account: %v", err)
	}

	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	req := journal.CreateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Non-posting account",
		Lines: []journal.JournalLineDto{
			{AccountID: headerAcc.ID, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}

	_, err = tc.jnlSvc.Create(ctx, userID, req)
	if err != journal.ErrNotPostingAccount {
		t.Fatalf("expected ErrNotPostingAccount, got %v", err)
	}
}

func TestService_Update_Success(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Create entry first
	createReq := journal.CreateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Original",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}
	created, err := tc.jnlSvc.Create(ctx, userID, createReq)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Update entry
	updateReq := journal.UpdateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Updated",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 200000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 200000},
		},
	}
	updated, err := tc.jnlSvc.Update(ctx, userID, created.ID, updateReq)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if updated.Description != "Updated" {
		t.Errorf("expected description 'Updated', got %s", updated.Description)
	}

	// Cleanup
	tc.jnlRepo.DeleteEntry(ctx, created.ID)
}

func TestService_Update_ErrNotFound(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	updateReq := journal.UpdateJournalRequest{
		Date:        time.Now().Format(time.RFC3339),
		Description: "Not found",
		Lines: []journal.JournalLineDto{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}

	_, err := tc.jnlSvc.Update(ctx, userID, 999999, updateReq)
	if err != journal.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestService_CreateBulk_Success(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Get account codes
	acc1, _ := tc.accRepo.FindByID(ctx, accID1)
	acc2, _ := tc.accRepo.FindByID(ctx, accID2)

	req := journal.CreateBulkJournalRequest{
		Entries: []journal.CreateJournalRequest{
			{
				Date:        time.Now().Format(time.RFC3339),
				Description: "Bulk entry 1",
				Lines: []journal.JournalLineDto{
					{AccountCode: acc1.Code, Debit: 100000, Credit: 0},
					{AccountCode: acc2.Code, Debit: 0, Credit: 100000},
				},
			},
			{
				Date:        time.Now().Format(time.RFC3339),
				Description: "Bulk entry 2",
				Lines: []journal.JournalLineDto{
					{AccountCode: acc1.Code, Debit: 200000, Credit: 0},
					{AccountCode: acc2.Code, Debit: 0, Credit: 200000},
				},
			},
		},
	}

	count, err := tc.jnlSvc.CreateBulk(ctx, userID, req)
	if err != nil {
		t.Fatalf("CreateBulk: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 entries created, got %d", count)
	}

	// Cleanup - find and delete the entries
	entries, _ := tc.jnlRepo.FindEntriesByUserID(ctx, userID, journal.EntryFilter{Limit: 100})
	for _, e := range entries {
		tc.jnlRepo.DeleteEntry(ctx, e.ID)
	}
}

func TestService_CreateBulk_ErrInvalidLines(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")

	acc1, _ := tc.accRepo.FindByID(ctx, accID1)

	req := journal.CreateBulkJournalRequest{
		Entries: []journal.CreateJournalRequest{
			{
				Date:        time.Now().Format(time.RFC3339),
				Description: "Too few lines",
				Lines: []journal.JournalLineDto{
					{AccountCode: acc1.Code, Debit: 100000, Credit: 0},
				},
			},
		},
	}

	_, err := tc.jnlSvc.CreateBulk(ctx, userID, req)
	if err != journal.ErrInvalidLines {
		t.Fatalf("expected ErrInvalidLines, got %v", err)
	}
}

func TestService_GetTotalSummary(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Create entry
	entry := &journal.JournalEntry{
		UserID:      userID,
		Date:        time.Now(),
		Description: "Summary test",
		Lines: []journal.JournalEntryLine{
			{AccountID: accID1, Debit: 100000, Credit: 0},
			{AccountID: accID2, Debit: 0, Credit: 100000},
		},
	}
	if err := tc.jnlRepo.CreateEntry(ctx, entry); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}

	// Get summary (entry uses time.Now, so it falls in the current month)
	summary, err := tc.jnlSvc.GetTotalSummary(ctx, userID, "")
	if err != nil {
		t.Fatalf("GetTotalSummary: %v", err)
	}

	if summary.TransactionCount < 1 {
		t.Errorf("expected at least 1 transaction, got %d", summary.TransactionCount)
	}

	// Cleanup
	tc.jnlRepo.DeleteEntry(ctx, entry.ID)
}
