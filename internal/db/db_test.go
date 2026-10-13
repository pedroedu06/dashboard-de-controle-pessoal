package db_test

import (
	"context"
	"testing"

	"dashboard/internal/db"
	"dashboard/internal/testutil"
)

func TestOpen_InvalidURL(t *testing.T) {
	_, err := db.Open(context.Background(), "isto não é uma url")
	if err == nil {
		t.Fatal("esperava erro para URL inválida")
	}
}

func TestMigrate_IsIdempotent(t *testing.T) {
	pool := testutil.Pool(t) // já aplica as migrations uma vez

	if err := db.Migrate(context.Background(), pool); err != nil {
		t.Fatalf("segunda execução de Migrate falhou: %v", err)
	}
}
