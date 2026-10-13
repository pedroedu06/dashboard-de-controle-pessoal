package db_test

import (
	"context"
	"testing"

	"dashboard/internal/db"
	"dashboard/internal/testutil"
)

func TestOpen_InvalidURL(t *testing.T) {
	_, err := db.Open(context.Background(), "not a url")
	if err == nil {
		t.Fatal("expected an error for an invalid URL")
	}
}

func TestMigrate_IsIdempotent(t *testing.T) {
	pool := testutil.Pool(t) // runs the migrations once

	if err := db.Migrate(context.Background(), pool); err != nil {
		t.Fatalf("second Migrate failed: %v", err)
	}
}
