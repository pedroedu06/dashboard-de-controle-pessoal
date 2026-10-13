package main

import (
	"context"
	"errors"
	"testing"

	"dashboard/internal/testutil"
)

func TestDBStatus_ReportsStartupError(t *testing.T) {
	app := &App{ctx: context.Background(), dbErr: errors.New("DATABASE_URL não definida")}

	got := app.DBStatus()
	if got.Connected || got.Error != "DATABASE_URL não definida" {
		t.Errorf("DBStatus = %+v", got)
	}
}

func TestDBStatus_NotInitialized(t *testing.T) {
	app := &App{ctx: context.Background()}

	if got := app.DBStatus(); got.Connected || got.Error == "" {
		t.Errorf("DBStatus = %+v, esperado erro", got)
	}
}

func TestDBStatus_Connected(t *testing.T) {
	app := &App{ctx: context.Background(), pool: testutil.Pool(t)}

	if got := app.DBStatus(); !got.Connected {
		t.Errorf("DBStatus = %+v, esperado conectado", got)
	}
}
