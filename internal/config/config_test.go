package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeEnv(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFrom_UsesEnvironmentVariable(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://from-env")

	cfg, err := LoadFrom()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DatabaseURL != "postgres://from-env" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
}

func TestLoadFrom_ReadsDotEnvFile(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	os.Unsetenv("DATABASE_URL")
	file := writeEnv(t, "DATABASE_URL=postgres://from-file\n")

	cfg, err := LoadFrom(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DatabaseURL != "postgres://from-file" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
}

func TestLoadFrom_EnvironmentWinsOverFile(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://from-env")
	file := writeEnv(t, "DATABASE_URL=postgres://from-file\n")

	cfg, err := LoadFrom(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DatabaseURL != "postgres://from-env" {
		t.Errorf("DatabaseURL = %q, want the environment value", cfg.DatabaseURL)
	}
}

func TestLoadFrom_IgnoresMissingFiles(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://from-env")

	if _, err := LoadFrom(filepath.Join(t.TempDir(), "missing.env")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadFrom_MissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	os.Unsetenv("DATABASE_URL")

	_, err := LoadFrom()
	if !errors.Is(err, ErrMissingDatabaseURL) {
		t.Fatalf("err = %v, want ErrMissingDatabaseURL", err)
	}
}
