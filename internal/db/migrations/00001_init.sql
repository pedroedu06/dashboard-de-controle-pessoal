-- +goose Up
-- Preferências do app (tema, módulo inicial etc.), em formato chave/valor.
CREATE TABLE settings (
    key        text        PRIMARY KEY,
    value      jsonb       NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE settings;
