// Package config carrega a configuração do app a partir do ambiente e de arquivos .env.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// AppName é o nome da pasta de configuração dentro de os.UserConfigDir().
const AppName = "dashboard"

// ErrMissingDatabaseURL indica que DATABASE_URL não foi encontrada em nenhuma fonte.
var ErrMissingDatabaseURL = errors.New("DATABASE_URL não definida: crie um arquivo .env (veja .env.example)")

// Config é a configuração do app.
type Config struct {
	DatabaseURL string
}

// Load lê a configuração, em ordem de prioridade:
//  1. variáveis de ambiente já definidas;
//  2. .env no diretório atual (desenvolvimento);
//  3. .env em os.UserConfigDir()/dashboard (app compilado).
//
// Arquivos .env nunca sobrescrevem variáveis já definidas no ambiente.
func Load() (Config, error) {
	return LoadFrom(candidateFiles()...)
}

// LoadFrom é como Load, mas recebe explicitamente os arquivos .env a tentar.
// Arquivos inexistentes são ignorados.
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
