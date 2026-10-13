// Package db concentra o acesso ao Postgres (Supabase): conexão, migrations e
// as queries geradas pelo sqlc (subpacote sqlc).
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Open cria o pool de conexões e confirma que o banco responde.
func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("configurando conexão com o banco: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("conectando ao banco: %w", err)
	}
	return pool, nil
}
