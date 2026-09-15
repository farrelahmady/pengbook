// Package journal_test contains integration tests for the journal module.
// These tests require an active PostgreSQL database (config read from .env).
package journal_test

import (
	"context"
	"fmt"
	"math/rand"
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
	accSvc := account.NewService(accRepo, txMgr)
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

// createAccount creates a test posting account and returns the ID.
func createAccount(t *testing.T, tc *testContext, userID int64, code string) int64 {
	t.Helper()
	ctx := context.Background()

	acc, err := tc.accSvc.Create(ctx, userID, account.CreateAccountRequest{
		Code: code,
		Name: fmt.Sprintf("Account %s", code),
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	return acc.ID
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

func TestRepository_GetSummaryByUserID(t *testing.T) {
	tc := setup(t)
	ctx := context.Background()
	userID := createUser(t, tc)
	accID1 := createAccount(t, tc, userID, "1.01.01.01")
	accID2 := createAccount(t, tc, userID, "4.01.01.01")

	// Create entries
	for i := 0; i < 3; i++ {
		entry := &journal.JournalEntry{
			UserID:      userID,
			Date:        time.Now(),
			Description: fmt.Sprintf("Entry %d", i),
			Lines: []journal.JournalEntryLine{
				{AccountID: accID1, Debit: 100000, Credit: 0},
				{AccountID: accID2, Debit: 0, Credit: 100000},
			},
		}
		if err := tc.jnlRepo.CreateEntry(ctx, entry); err != nil {
			t.Fatalf("CreateEntry: %v", err)
		}
		defer tc.jnlRepo.DeleteEntry(ctx, entry.ID)
	}

	// Get summary
	summary, err := tc.jnlRepo.GetSummaryByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetSummaryByUserID: %v", err)
	}

	if summary.TransactionCount < 3 {
		t.Errorf("expected at least 3 transactions, got %d", summary.TransactionCount)
	}
	if summary.TotalDebit < 300000 {
		t.Errorf("expected total debit >= 300000, got %f", summary.TotalDebit)
	}
	if summary.TotalCredit < 300000 {
		t.Errorf("expected total credit >= 300000, got %f", summary.TotalCredit)
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
	headerAcc, err := tc.accSvc.Create(ctx, userID, account.CreateAccountRequest{
		Code: "1.01.00.00",
		Name: "Header Account",
	})
	if err != nil {
		t.Fatalf("create header account: %v", err)
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

	// Get summary
	summary, err := tc.jnlSvc.GetTotalSummary(ctx, userID)
	if err != nil {
		t.Fatalf("GetTotalSummary: %v", err)
	}

	if summary.TransactionCount < 1 {
		t.Errorf("expected at least 1 transaction, got %d", summary.TransactionCount)
	}

	// Cleanup
	tc.jnlRepo.DeleteEntry(ctx, entry.ID)
}
