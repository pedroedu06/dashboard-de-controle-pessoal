package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"dashboard/internal/config"
	"dashboard/internal/db"
	"dashboard/internal/db/sqlc"
	"dashboard/internal/tasks"
)

// App holds the state shared by the Wails bindings.
type App struct {
	ctx   context.Context
	pool  *pgxpool.Pool
	dbErr error
	tasks *tasks.Service // nil while the database is unavailable
}

// DBStatus is the database connection state sent to the frontend.
type DBStatus struct {
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// A database failure must not crash the app; it is reported via DBStatus.
	a.pool, a.dbErr = connect(ctx)
	if a.dbErr == nil {
		a.tasks = tasks.NewService(sqlc.New(a.pool))
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.pool != nil {
		a.pool.Close()
	}
}

// DBStatus reports whether the app can reach the database.
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
