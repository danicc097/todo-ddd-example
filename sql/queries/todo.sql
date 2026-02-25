-- name: UpsertTodo :one
INSERT INTO todos(id, title, status, created_at, updated_at, workspace_id, due_date, recurrence_interval, recurrence_amount, last_completed_at, deleted_at)
  VALUES ($1, $2, $3, $4, $4, $5, $6, $7, $8, $9, NULL)
ON CONFLICT (id)
  DO UPDATE SET
    title = EXCLUDED.title,
    status = EXCLUDED.status,
    updated_at = NOW(),
    due_date = EXCLUDED.due_date,
    recurrence_interval = EXCLUDED.recurrence_interval,
    recurrence_amount = EXCLUDED.recurrence_amount,
    last_completed_at = EXCLUDED.last_completed_at,
    deleted_at = NULL
  RETURNING
    *;

-- name: GetTodoAggregateByID :one
SELECT
  t.*,
  COALESCE(array_remove(array_agg(DISTINCT tt.tag_id), NULL), '{}')::uuid[] AS tags,
  COALESCE((
    SELECT
      json_agg(fs.*)
    FROM todo_focus_sessions fs
    WHERE
      fs.todo_id = t.id), '[]'::json) AS focus_sessions
FROM
  todos t
  LEFT JOIN todo_tags tt ON t.id = tt.todo_id
WHERE
  t.id = $1
  AND t.deleted_at IS NULL
GROUP BY
  t.id;

-- name: GetTodoReadModelByID :one
SELECT
  t.*,
  COALESCE(array_remove(array_agg(DISTINCT tt.tag_id), NULL), '{}')::uuid[] AS tags,
  COALESCE((
    SELECT
      json_agg(fs.*)
    FROM todo_focus_sessions fs
    WHERE
      fs.todo_id = t.id), '[]'::json) AS focus_sessions
FROM
  todos t
  LEFT JOIN todo_tags tt ON t.id = tt.todo_id
WHERE
  t.id = $1
  AND t.deleted_at IS NULL
GROUP BY
  t.id;

-- name: ListTodosByWorkspaceID :many
SELECT
  t.*,
  COALESCE(array_remove(array_agg(DISTINCT tt.tag_id), NULL), '{}')::uuid[] AS tags,
  COALESCE((
    SELECT
      json_agg(fs.*)
    FROM todo_focus_sessions fs
    WHERE
      fs.todo_id = t.id), '[]'::json) AS focus_sessions
FROM
  todos t
  LEFT JOIN todo_tags tt ON t.id = tt.todo_id
WHERE
  t.workspace_id = $1
  AND t.deleted_at IS NULL
GROUP BY
  t.id
ORDER BY
  t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: BulkAddTagsToTodo :exec
INSERT INTO todo_tags(todo_id, tag_id)
SELECT
  UNNEST(sqlc.arg(todo_ids)::uuid[]),
  UNNEST(sqlc.arg(tag_ids)::uuid[])
ON CONFLICT
  DO NOTHING;

-- name: RemoveMissingTagsFromTodo :exec
DELETE FROM todo_tags
WHERE todo_id = $1
  AND NOT (tag_id = ANY (sqlc.arg(tags)::uuid[]));

-- name: DeleteTodo :exec
UPDATE
  todos
SET
  deleted_at = NOW()
WHERE
  id = $1;

-- name: UpsertFocusSession :exec
INSERT INTO todo_focus_sessions(id, todo_id, start_time, end_time)
  VALUES ($1, $2, $3, $4)
ON CONFLICT (id)
  DO UPDATE SET
    end_time = EXCLUDED.end_time;

-- name: BulkUpsertFocusSessions :exec
INSERT INTO todo_focus_sessions(id, todo_id, user_id, start_time, end_time)
SELECT
  UNNEST(sqlc.arg(ids)::uuid[]),
  UNNEST(sqlc.arg(todo_ids)::uuid[]),
  UNNEST(sqlc.arg(user_ids)::uuid[]),
  UNNEST(sqlc.arg(start_times)::timestamptz[]),
  NULLIF(UNNEST(sqlc.arg(end_times)::timestamptz[]), '0001-01-01 00:00:00+00'::timestamptz)
ON CONFLICT (id)
  DO UPDATE SET
    end_time = EXCLUDED.end_time;

-- name: RemoveMissingFocusSessionsFromTodo :exec
DELETE FROM todo_focus_sessions
WHERE todo_id = $1
  AND NOT (id = ANY (sqlc.arg(session_ids)::uuid[]));

