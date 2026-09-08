package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"pengbook/api/internal/config"
)

// NewTestPool creates a database connection pool for tests (integration test).
//
// Steps:
//  1. Load the .env file from the repo root (searched upwards from the working
//     directory, because `go test` runs from the package directory).
//  2. Read config and open a PostgreSQL pool (same as production).
//  3. Run migrations so the schema tables are available.
//  4. Register pool.Close() via t.Cleanup so it is closed automatically.
//
// Requires an active PostgreSQL matching the .env (e.g. database `pengbook`).
func NewTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	_ = godotenv.Load(findEnvFile())

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	pool, err := NewPostgres(cfg.Database)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := RunMigrations(ctx, pool); err != nil {
		pool.Close()
		t.Fatalf("run migrations: %v", err)
	}

	t.Cleanup(pool.Close)

	return pool
}

// findEnvFile searches for a .env file starting from the current working
// directory, walking up to parent directories until it finds one (or reaches
// the root).
//
// During `go test`, the working directory is the package folder (e.g.
// internal/module/user) while .env lives at the repo root.
func findEnvFile() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		p := filepath.Join(dir, ".env")
		if _, err := os.Stat(p); err == nil {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
