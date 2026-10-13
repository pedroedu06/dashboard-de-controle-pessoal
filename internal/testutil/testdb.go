// Package testutil provides helpers for Postgres integration tests.
package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"dashboard/internal/db"
	"dashboard/internal/db/sqlc"
)

const EnvTestDatabaseURL = "TEST_DATABASE_URL"

// DatabaseURL returns the test database URL, or skips the test when it is unset.
func DatabaseURL(t testing.TB) string {
	t.Helper()
	url := os.Getenv(EnvTestDatabaseURL)
	if url == "" {
		t.Skipf("%s not set: start the test database with `go tool task db:test:up` and run `go tool task test`", EnvTestDatabaseURL)
	}
	if url == os.Getenv("DATABASE_URL") {
		t.Fatalf("%s equals DATABASE_URL: tests must never run against the real database", EnvTestDatabaseURL)
	}
	return url
}

// Pool opens a migrated test database pool, closed when the test ends.
func Pool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pool, err := db.Open(ctx, DatabaseURL(t))
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}

// Queries returns sqlc queries bound to a transaction that is rolled back
// when the test ends, so tests never see each other's data.
func Queries(t testing.TB) *sqlc.Queries {
	t.Helper()
	ctx := context.Background()

	tx, err := Pool(t).Begin(ctx)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}
	// Cleanups run LIFO, so this rollback happens before the pool closes.
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	return sqlc.New(tx)
}
