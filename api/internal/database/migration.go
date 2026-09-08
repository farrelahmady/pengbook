package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"pengbook/api/migrations"
)

// RunMigrations applies all pending goose migrations.
// Uses the embedded migration files from the migrations package.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	db, err := getSQLDB(pool)
	if err != nil {
		return fmt.Errorf("open sql db: %w", err)
	}
	defer db.Close()

	// Use embedded filesystem
	fsys, err := fs.Sub(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("sub fs: %w", err)
	}

	goose.SetBaseFS(fsys)
	goose.SetDialect("postgres")

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

// RunMigrationsDown rolls back the most recent goose migration.
func RunMigrationsDown(ctx context.Context, pool *pgxpool.Pool) error {
	db, err := getSQLDB(pool)
	if err != nil {
		return fmt.Errorf("open sql db: %w", err)
	}
	defer db.Close()

	fsys, err := fs.Sub(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("sub fs: %w", err)
	}

	goose.SetBaseFS(fsys)
	goose.SetDialect("postgres")

	if err := goose.Down(db, "."); err != nil {
		return fmt.Errorf("goose down: %w", err)
	}

	return nil
}

// RunMigrationsStatus prints the status of all migrations.
func RunMigrationsStatus(ctx context.Context, pool *pgxpool.Pool) error {
	db, err := getSQLDB(pool)
	if err != nil {
		return fmt.Errorf("open sql db: %w", err)
	}
	defer db.Close()

	fsys, err := fs.Sub(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("sub fs: %w", err)
	}

	goose.SetBaseFS(fsys)
	goose.SetDialect("postgres")

	if err := goose.Status(db, "."); err != nil {
		return fmt.Errorf("goose status: %w", err)
	}

	return nil
}

// getSQLDB creates a temporary *sql.DB from the pgxpool for goose operations.
func getSQLDB(pool *pgxpool.Pool) (*sql.DB, error) {
	connStr := stdlib.RegisterConnConfig(pool.Config().ConnConfig)
	return sql.Open("pgx", connStr)
}
