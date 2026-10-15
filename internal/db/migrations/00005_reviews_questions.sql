-- +goose Up
-- SM-2 spaced repetition state; next_review_on NULL means not scheduled.
ALTER TABLE topics
    ADD COLUMN next_review_on date,
    ADD COLUMN interval_days  int              NOT NULL DEFAULT 0 CHECK (interval_days >= 0),
    ADD COLUMN ease_factor    double precision NOT NULL DEFAULT 2.5 CHECK (ease_factor >= 1.3),
    ADD COLUMN repetitions    int              NOT NULL DEFAULT 0 CHECK (repetitions >= 0);

CREATE INDEX topics_next_review_idx ON topics (next_review_on);

CREATE TABLE topic_reviews (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    topic_id      bigint      NOT NULL REFERENCES topics ON DELETE CASCADE,
    reviewed_at   timestamptz NOT NULL,
    grade         smallint    NOT NULL CHECK (grade BETWEEN 0 AND 5),
    interval_days int         NOT NULL
);

CREATE TABLE question_logs (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    subject_id bigint      NOT NULL REFERENCES subjects,
    topic_id   bigint      REFERENCES topics ON DELETE SET NULL,
    done_on    date        NOT NULL,
    total      int         NOT NULL CHECK (total BETWEEN 1 AND 1000),
    correct    int         NOT NULL CHECK (correct BETWEEN 0 AND total),
    source     text        CHECK (source IS NULL OR btrim(source) <> ''),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX question_logs_done_on_idx ON question_logs (done_on);

-- +goose Down
DROP TABLE question_logs;
DROP TABLE topic_reviews;
DROP INDEX topics_next_review_idx;
ALTER TABLE topics
    DROP COLUMN repetitions,
    DROP COLUMN ease_factor,
    DROP COLUMN interval_days,
    DROP COLUMN next_review_on;
