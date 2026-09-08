// Package user_test contains integration tests for the user service that
// require an active PostgreSQL (config read from .env).
//
// The tests demonstrate the transaction-manager pattern:
//   - TestCreate_WithTestTx_RollbackNotCommitted: the TEST owns the
//     transaction (Begin + Rollback at the end); the service never commits.
//   - TestCreate_ServiceCommits: without an injected transaction, the service
//     begins and commits its own transaction.
package user_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengbook/api/internal/database"
	"pengbook/api/internal/infrastructure/postgres"
	"pengbook/api/internal/module/user"
)

// setup builds the pool, repository, and service used by the tests.
// The pool is created via database.NewTestPool (loads .env, runs migrations,
// auto-closes via t.Cleanup).
func setup(t *testing.T) (*pgxpool.Pool, user.Service, user.Repository) {
	t.Helper()

	pool := database.NewTestPool(t)
	repo := postgres.NewUserRepository(pool)
	svc := user.NewService(repo, postgres.NewTxManager(pool))

	return pool, svc, repo
}

// createReq returns a CreateUserRequest with a unique email per call.
func createReq() user.CreateUserRequest {
	return user.CreateUserRequest{
		Name:     "Test User",
		Email:    fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		Password: "password123",
	}
}

// TestCreate_WithTestTx_RollbackNotCommitted verifies that when a test begins
// its own transaction, injects it into the context, and rolls back at the end,
// the service does NOT commit anything.
func TestCreate_WithTestTx_RollbackNotCommitted(t *testing.T) {
	pool, svc, repo := setup(t)
	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	// Run the service with the test's transaction injected into the context.
	txCtx := database.WithTx(ctx, tx)
	created, err := svc.Create(txCtx, createReq())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero user id")
	}

	// Still inside the transaction → the user is visible via the same tx.
	got, err := repo.FindByID(txCtx, created.ID)
	if err != nil {
		t.Fatalf("find inside tx: %v", err)
	}
	if got == nil {
		t.Fatal("expected user visible inside tx before rollback")
	}

	// The service did NOT commit; the test rolls back.
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	// Verify via the pool (a different connection): nothing was persisted.
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM users WHERE id = $1", created.ID).Scan(&count); err != nil {
		t.Fatalf("query users after rollback: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected user rolled back, found %d row(s)", count)
	}

	if err := pool.QueryRow(ctx, "SELECT count(*) FROM user_audit_logs WHERE user_id = $1", created.ID).Scan(&count); err != nil {
		t.Fatalf("query audit after rollback: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected audit log rolled back, found %d row(s)", count)
	}
}

// TestCreate_ServiceCommits verifies the normal path: without an injected
// transaction, the service begins and commits its own transaction.
func TestCreate_ServiceCommits(t *testing.T) {
	pool, svc, repo := setup(t)
	ctx := context.Background()

	req := createReq()
	created, err := svc.Create(ctx, req)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	u, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if u == nil {
		t.Fatal("expected committed user")
	}
	if u.Name != req.Name || u.Email != req.Email {
		t.Fatalf("unexpected user: %+v", u)
	}

	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM user_audit_logs WHERE user_id = $1", created.ID).Scan(&count); err != nil {
		t.Fatalf("query audit: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 audit log committed, found %d", count)
	}

	if _, err := pool.Exec(ctx, "DELETE FROM users WHERE id = $1", created.ID); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
}

// TestGetByID_NotFound verifies that an unknown id returns ErrNotFound.
func TestGetByID_NotFound(t *testing.T) {
	_, svc, _ := setup(t)
	ctx := context.Background()

	if _, err := svc.GetByID(ctx, -1); err != user.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
