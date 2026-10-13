// Package testutil reúne helpers para testes de integração com o Postgres.
package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"dashboard/internal/db"
	"dashboard/internal/db/sqlc"
)

// EnvTestDatabaseURL é a variável com a URL do banco usado pelos testes.
const EnvTestDatabaseURL = "TEST_DATABASE_URL"

// DatabaseURL devolve a URL do banco de testes ou pula o teste se ela não
// estiver definida (por exemplo, sem o Docker rodando).
func DatabaseURL(t testing.TB) string {
	t.Helper()
	url := os.Getenv(EnvTestDatabaseURL)
	if url == "" {
		t.Skipf("%s não definida: suba o banco de testes com `go tool task db:test:up` e rode via `go tool task test`", EnvTestDatabaseURL)
	}
	if url == os.Getenv("DATABASE_URL") {
		t.Fatalf("%s é igual a DATABASE_URL: os testes não podem rodar no banco real", EnvTestDatabaseURL)
	}
	return url
}

// Pool abre um pool no banco de testes com todas as migrations aplicadas.
// O pool é fechado ao fim do teste.
func Pool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pool, err := db.Open(ctx, DatabaseURL(t))
	if err != nil {
		t.Fatalf("abrindo banco de testes: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("aplicando migrations: %v", err)
	}
	return pool
}

// Queries devolve as queries do sqlc dentro de uma transação que sofre
// rollback ao fim do teste, isolando cada teste dos demais.
func Queries(t testing.TB) *sqlc.Queries {
	t.Helper()
	ctx := context.Background()

	tx, err := Pool(t).Begin(ctx)
	if err != nil {
		t.Fatalf("iniciando transação: %v", err)
	}
	// Registrado depois do Close do pool, então roda antes dele (ordem LIFO).
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	return sqlc.New(tx)
}
