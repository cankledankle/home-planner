package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/cankledankle/home-planner/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// newTestStore creates a Store backed by the TEST_DATABASE_URL database.
// It skips the test if TEST_DATABASE_URL is not set.
// The pool is closed automatically at the end of the test.
func newTestStore(t *testing.T) *db.Store {
	t.Helper()

	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		t.Fatalf("failed to create test pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	// Run migrations so the schema is up to date.
	t.Setenv("DATABASE_URL", connStr)
	t.Setenv("MIGRATIONS_PATH", "file://../../migrations")
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db.NewStore(pool)
}
