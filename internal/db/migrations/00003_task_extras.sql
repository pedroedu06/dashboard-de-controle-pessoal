-- +goose Up
CREATE TYPE task_repeat AS ENUM ('none', 'daily', 'weekly');

ALTER TABLE tasks
    ADD COLUMN category        text CHECK (category IS NULL OR btrim(category) <> ''),
    ADD COLUMN repeat          task_repeat NOT NULL DEFAULT 'none',
    ADD COLUMN repeat_weekdays smallint[]  NOT NULL DEFAULT '{}', -- 0 = Sunday .. 6 = Saturday
    ADD COLUMN repeat_until    date,
    ADD CONSTRAINT tasks_weekly_has_weekdays CHECK (repeat <> 'weekly' OR cardinality(repeat_weekdays) > 0);

CREATE INDEX tasks_category_idx ON tasks (lower(category));

-- Checklist items. done_at is used by non-recurring tasks only.
CREATE TABLE task_items (
    id       bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id  bigint      NOT NULL REFERENCES tasks ON DELETE CASCADE,
    title    text        NOT NULL CHECK (btrim(title) <> ''),
    position int         NOT NULL,
    done_at  timestamptz
);

CREATE INDEX task_items_task_idx ON task_items (task_id, position);

-- A row means the recurring task was completed on that day.
CREATE TABLE task_occurrences (
    task_id   bigint      NOT NULL REFERENCES tasks ON DELETE CASCADE,
    occurs_on date        NOT NULL,
    done_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (task_id, occurs_on)
);

-- A row means the item was checked on that occurrence of a recurring task.
CREATE TABLE task_item_checks (
    item_id   bigint NOT NULL REFERENCES task_items ON DELETE CASCADE,
    occurs_on date   NOT NULL,
    PRIMARY KEY (item_id, occurs_on)
);

-- +goose Down
DROP TABLE task_item_checks;
DROP TABLE task_occurrences;
DROP TABLE task_items;
DROP INDEX tasks_category_idx;
ALTER TABLE tasks
    DROP CONSTRAINT tasks_weekly_has_weekdays,
    DROP COLUMN repeat_until,
    DROP COLUMN repeat_weekdays,
    DROP COLUMN repeat,
    DROP COLUMN category;
DROP TYPE task_repeat;
