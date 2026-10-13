package db_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"dashboard/internal/db/sqlc"
	"dashboard/internal/testutil"
)

func TestSettings_UpsertAndGet(t *testing.T) {
	ctx := context.Background()
	q := testutil.Queries(t)

	created, err := q.UpsertSetting(ctx, sqlc.UpsertSettingParams{
		Key:   "theme",
		Value: json.RawMessage(`"dark"`),
	})
	if err != nil {
		t.Fatalf("UpsertSetting: %v", err)
	}

	updated, err := q.UpsertSetting(ctx, sqlc.UpsertSettingParams{
		Key:   "theme",
		Value: json.RawMessage(`"light"`),
	})
	if err != nil {
		t.Fatalf("UpsertSetting (update): %v", err)
	}
	if updated.UpdatedAt.Before(created.UpdatedAt) {
		t.Errorf("updated_at não avançou: %v -> %v", created.UpdatedAt, updated.UpdatedAt)
	}

	got, err := q.GetSetting(ctx, "theme")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if string(got.Value) != `"light"` {
		t.Errorf("Value = %s, esperado \"light\"", got.Value)
	}
}

func TestSettings_GetMissingKey(t *testing.T) {
	q := testutil.Queries(t)

	_, err := q.GetSetting(context.Background(), "nao-existe")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("err = %v, esperado pgx.ErrNoRows", err)
	}
}

func TestSettings_ListOrderedByKey(t *testing.T) {
	ctx := context.Background()
	q := testutil.Queries(t)

	for _, key := range []string{"b", "a", "c"} {
		if _, err := q.UpsertSetting(ctx, sqlc.UpsertSettingParams{Key: key, Value: json.RawMessage(`true`)}); err != nil {
			t.Fatalf("UpsertSetting(%s): %v", key, err)
		}
	}

	list, err := q.ListSettings(ctx)
	if err != nil {
		t.Fatalf("ListSettings: %v", err)
	}
	var keys []string
	for _, s := range list {
		keys = append(keys, s.Key)
	}
	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("keys = %v, esperado [a b c]", keys)
	}
}
