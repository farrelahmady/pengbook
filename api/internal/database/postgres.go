// Package database contains database infrastructure: connection pool,
// query interface (DBTX), transaction manager, and migration runner.
package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"pengbook/api/internal/config"
)

// DBTX is the minimal query interface satisfied by both *pgxpool.Pool
// (regular connection) and pgx.Tx (transaction).
//
// Repositories write queries against DBTX so the same query set can be used
// inside or outside a transaction without knowing the underlying type.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// NewPostgres builds a PostgreSQL connection pool from the database config.
//
// The DSN is assembled with url.UserPassword so passwords containing special
// characters (e.g. "@", ":") are safely escaped. The pool is configured with
// MaxConns=10 & MinConns=1, then verified via Ping (pool closed on failure).
func NewPostgres(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	query := url.Values{}
	query.Set("sslmode", cfg.SSLMode)

	dsn := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:     cfg.Name,
		RawQuery: query.Encode(),
	}).String()

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	poolCfg.MaxConns = 10
	poolCfg.MinConns = 1

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
