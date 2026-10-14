-- name: CreateTask :one
INSERT INTO tasks (title, urgency, starts_at, ends_at, category, repeat, repeat_weekdays, repeat_until)
VALUES (@title, @urgency, @starts_at, @ends_at, @category, @repeat, @repeat_weekdays, @repeat_until)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = @id;

-- name: UpdateTask :one
UPDATE tasks
SET title = @title,
    urgency = @urgency,
    starts_at = @starts_at,
    ends_at = @ends_at,
    category = @category,
    repeat = @repeat,
    repeat_weekdays = @repeat_weekdays,
    repeat_until = @repeat_until,
    updated_at = now()
WHERE id = @id
RETURNING *;

-- name: DeleteTask :execrows
DELETE FROM tasks
WHERE id = @id;

-- name: SetTaskDone :one
UPDATE tasks
SET done_at = CASE WHEN @done::boolean THEN coalesce(done_at, now()) END,
    updated_at = now()
WHERE id = @id
RETURNING *;

-- Agenda: recurring tasks, everything ending at or after @since, and anything still not done.
-- name: ListAgendaTasks :many
SELECT * FROM tasks
WHERE repeat <> 'none' OR ends_at >= @since OR done_at IS NULL
ORDER BY starts_at, id;

-- name: FindCategory :one
SELECT category::text FROM tasks
WHERE lower(category) = lower(@name::text)
LIMIT 1;

-- name: ListCategories :many
SELECT min(category)::text AS category FROM tasks
WHERE category IS NOT NULL
GROUP BY lower(category)
ORDER BY lower(category);

-- name: ListItemsByTasks :many
SELECT * FROM task_items
WHERE task_id = ANY(@task_ids::bigint[])
ORDER BY task_id, position, id;

-- name: GetItem :one
SELECT * FROM task_items
WHERE id = @id;

-- name: InsertItem :one
INSERT INTO task_items (task_id, title, position)
VALUES (@task_id, @title, @position)
RETURNING *;

-- name: UpdateItem :execrows
UPDATE task_items
SET title = @title,
    position = @position
WHERE id = @id AND task_id = @task_id;

-- name: DeleteItemsNotIn :exec
DELETE FROM task_items
WHERE task_id = @task_id AND NOT (id = ANY(@keep_ids::bigint[]));

-- name: SetItemDone :exec
UPDATE task_items
SET done_at = CASE WHEN @done::boolean THEN coalesce(done_at, now()) END
WHERE id = @id;

-- name: ListOccurrences :many
SELECT * FROM task_occurrences
WHERE task_id = ANY(@task_ids::bigint[])
  AND occurs_on BETWEEN @from_day AND @to_day;

-- name: UpsertOccurrence :exec
INSERT INTO task_occurrences (task_id, occurs_on)
VALUES (@task_id, @occurs_on)
ON CONFLICT DO NOTHING;

-- name: DeleteOccurrence :exec
DELETE FROM task_occurrences
WHERE task_id = @task_id AND occurs_on = @occurs_on;

-- name: ListItemChecks :many
SELECT * FROM task_item_checks
WHERE item_id = ANY(@item_ids::bigint[])
  AND occurs_on BETWEEN @from_day AND @to_day;

-- name: InsertItemCheck :exec
INSERT INTO task_item_checks (item_id, occurs_on)
VALUES (@item_id, @occurs_on)
ON CONFLICT DO NOTHING;

-- name: DeleteItemCheck :exec
DELETE FROM task_item_checks
WHERE item_id = @item_id AND occurs_on = @occurs_on;
