-- +goose Up
CREATE TYPE task_urgency AS ENUM ('low', 'medium', 'high');

CREATE TABLE tasks (
    id         bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title      text         NOT NULL CHECK (btrim(title) <> ''),
    urgency    task_urgency NOT NULL DEFAULT 'medium',
    starts_at  timestamptz  NOT NULL,
    ends_at    timestamptz  NOT NULL,
    done_at    timestamptz,
    created_at timestamptz  NOT NULL DEFAULT now(),
    updated_at timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT tasks_ends_after_start CHECK (ends_at > starts_at)
);

CREATE INDEX tasks_ends_at_idx ON tasks (ends_at);

-- +goose Down
DROP TABLE tasks;
DROP TYPE task_urgency;
