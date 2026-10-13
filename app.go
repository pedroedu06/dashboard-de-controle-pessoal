package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"dashboard/internal/config"
	"dashboard/internal/db"
)

// App struct
type App struct {
	ctx   context.Context
	pool  *pgxpool.Pool
	dbErr error
}

// DBStatus é o estado da conexão com o banco, exposto ao frontend.
type DBStatus struct {
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Falhas de banco não derrubam o app: ficam visíveis via DBStatus.
	a.pool, a.dbErr = connect(ctx)
}

// shutdown is called when the app is closing.
func (a *App) shutdown(ctx context.Context) {
	if a.pool != nil {
		a.pool.Close()
	}
}

// DBStatus informa se o app está conectado ao banco.
func (a *App) DBStatus() DBStatus {
	if a.dbErr != nil {
		return DBStatus{Error: a.dbErr.Error()}
	}
	if a.pool == nil {
		return DBStatus{Error: "banco não inicializado"}
	}
	if err := a.pool.Ping(a.ctx); err != nil {
		return DBStatus{Error: err.Error()}
	}
	return DBStatus{Connected: true}
}

func connect(ctx context.Context) (*pgxpool.Pool, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
