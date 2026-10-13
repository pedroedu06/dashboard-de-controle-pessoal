// Package config loads the app configuration from the environment and .env files.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// AppName is the config folder name inside os.UserConfigDir().
const AppName = "dashboard"

var ErrMissingDatabaseURL = errors.New("DATABASE_URL não definida: crie um arquivo .env (veja docs/ambiente.md)")

type Config struct {
	DatabaseURL string
}

// Load reads the configuration, in order of precedence:
//  1. variables already set in the environment;
//  2. .env in the working directory (development);
//  3. .env in os.UserConfigDir()/dashboard (built app).
//
// .env files never override variables already set.
func Load() (Config, error) {
	return LoadFrom(candidateFiles()...)
}

// LoadFrom is like Load but with explicit .env files. Missing files are skipped.
func LoadFrom(files ...string) (Config, error) {
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			continue
		}
		if err := godotenv.Load(f); err != nil {
			return Config{}, fmt.Errorf("lendo %s: %w", f, err)
		}
	}

	cfg := Config{DatabaseURL: os.Getenv("DATABASE_URL")}
	if cfg.DatabaseURL == "" {
		return Config{}, ErrMissingDatabaseURL
	}
	return cfg, nil
}

func candidateFiles() []string {
	files := []string{".env"}
	if dir, err := os.UserConfigDir(); err == nil {
		files = append(files, filepath.Join(dir, AppName, ".env"))
	}
	return files
}
