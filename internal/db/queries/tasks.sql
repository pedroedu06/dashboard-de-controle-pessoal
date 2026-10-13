-- name: CreateTask :one
INSERT INTO tasks (title, urgency, starts_at, ends_at)
VALUES (@title, @urgency, @starts_at, @ends_at)
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

-- Agenda: everything ending at or after @since, plus anything still not done.
-- name: ListAgendaTasks :many
SELECT * FROM tasks
WHERE ends_at >= @since OR done_at IS NULL
ORDER BY starts_at, id;
