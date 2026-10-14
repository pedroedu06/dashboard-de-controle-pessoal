-- name: ListSubjects :many
SELECT * FROM subjects
WHERE @include_archived::boolean OR archived_at IS NULL
ORDER BY lower(name), id;

-- name: GetSubject :one
SELECT * FROM subjects
WHERE id = @id;

-- name: CreateSubject :one
INSERT INTO subjects (name)
VALUES (@name)
RETURNING *;

-- name: RenameSubject :execrows
UPDATE subjects
SET name = @name
WHERE id = @id;

-- name: SetSubjectArchived :execrows
UPDATE subjects
SET archived_at = CASE WHEN @archived::boolean THEN coalesce(archived_at, now()) END
WHERE id = @id;

-- name: ListTopicsBySubjects :many
SELECT * FROM topics
WHERE subject_id = ANY(@subject_ids::bigint[])
ORDER BY subject_id, id;

-- name: GetTopic :one
SELECT * FROM topics
WHERE id = @id;

-- name: CreateTopic :one
INSERT INTO topics (subject_id, title)
VALUES (@subject_id, @title)
RETURNING *;

-- name: RenameTopic :execrows
UPDATE topics
SET title = @title
WHERE id = @id;

-- name: SetTopicStatus :execrows
UPDATE topics
SET status = @status
WHERE id = @id;

-- name: DeleteTopic :execrows
DELETE FROM topics
WHERE id = @id;

-- name: GetActiveSession :one
SELECT s.*, sub.name AS subject_name, t.title AS topic_title
FROM study_sessions s
JOIN subjects sub ON sub.id = s.subject_id
LEFT JOIN topics t ON t.id = s.topic_id
WHERE s.ended_at IS NULL;

-- name: StartSession :one
INSERT INTO study_sessions (subject_id, topic_id, started_at)
VALUES (@subject_id, @topic_id, @started_at)
RETURNING id;

-- name: UpdateSessionTimer :exec
UPDATE study_sessions
SET paused_at = @paused_at,
    paused_seconds = @paused_seconds,
    ended_at = @ended_at
WHERE id = @id;

-- name: DeleteSession :execrows
DELETE FROM study_sessions
WHERE id = @id;

-- Finished sessions that started in [@from_time, @to_time).
-- name: ListSessionsBetween :many
SELECT s.*, sub.name AS subject_name, t.title AS topic_title
FROM study_sessions s
JOIN subjects sub ON sub.id = s.subject_id
LEFT JOIN topics t ON t.id = s.topic_id
WHERE s.ended_at IS NOT NULL
  AND s.started_at >= @from_time AND s.started_at < @to_time
ORDER BY s.started_at DESC, s.id DESC;

-- Goals of active subjects, plus overall goals.
-- name: ListGoals :many
SELECT g.*, sub.name AS subject_name
FROM study_goals g
LEFT JOIN subjects sub ON sub.id = g.subject_id
WHERE g.subject_id IS NULL OR sub.archived_at IS NULL
ORDER BY g.subject_id NULLS FIRST, lower(sub.name), g.id;

-- name: CreateGoal :one
INSERT INTO study_goals (subject_id, target_minutes)
VALUES (@subject_id, @target_minutes)
RETURNING id;

-- name: UpdateGoalTarget :execrows
UPDATE study_goals
SET target_minutes = @target_minutes
WHERE id = @id;

-- name: DeleteGoal :execrows
DELETE FROM study_goals
WHERE id = @id;
