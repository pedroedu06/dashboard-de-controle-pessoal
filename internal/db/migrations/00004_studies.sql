-- +goose Up
CREATE TABLE subjects (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        text        NOT NULL CHECK (btrim(name) <> ''),
    archived_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- Active subject names are unique, ignoring case; archived ones may repeat.
CREATE UNIQUE INDEX subjects_active_name_idx ON subjects (lower(name)) WHERE archived_at IS NULL;

CREATE TYPE topic_status AS ENUM ('not_seen', 'studying', 'reviewed');

CREATE TABLE topics (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    subject_id bigint       NOT NULL REFERENCES subjects ON DELETE CASCADE,
    title      text         NOT NULL CHECK (btrim(title) <> ''),
    status     topic_status NOT NULL DEFAULT 'not_seen',
    created_at timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX topics_subject_idx ON topics (subject_id, id);

-- A session with ended_at NULL is the running timer.
CREATE TABLE study_sessions (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    subject_id     bigint      NOT NULL REFERENCES subjects,
    topic_id       bigint      REFERENCES topics ON DELETE SET NULL,
    started_at     timestamptz NOT NULL,
    ended_at       timestamptz,
    paused_at      timestamptz,
    paused_seconds int         NOT NULL DEFAULT 0 CHECK (paused_seconds >= 0),
    CHECK (ended_at IS NULL OR ended_at >= started_at)
);

-- At most one running timer.
CREATE UNIQUE INDEX study_sessions_one_running_idx ON study_sessions ((true)) WHERE ended_at IS NULL;
CREATE INDEX study_sessions_started_idx ON study_sessions (started_at);

-- subject_id NULL means a goal over all subjects.
CREATE TABLE study_goals (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    subject_id     bigint      REFERENCES subjects,
    target_minutes int         NOT NULL CHECK (target_minutes > 0 AND target_minutes <= 10080),
    created_at     timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE study_goals;
DROP TABLE study_sessions;
DROP TABLE topics;
DROP TYPE topic_status;
DROP TABLE subjects;
