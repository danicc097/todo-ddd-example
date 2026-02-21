-- name: UpsertTodo :one
INSERT INTO todos(id, title, status, created_at, updated_at, workspace_id)
  VALUES ($1, $2, $3, $4, $4, $5)
ON CONFLICT (id)
  DO UPDATE SET
    title = EXCLUDED.title,
    status = EXCLUDED.status,
    updated_at = NOW()
  RETURNING
    id,
    title,
    status,
    created_at,
    workspace_id;

-- name: GetTodoByID :one
SELECT
  t.id,
  t.title,
  t.status,
  t.created_at,
  t.workspace_id,
  COALESCE(array_remove(array_agg(tt.tag_id), NULL), '{}')::uuid[] AS tags
FROM
  todos t
  LEFT JOIN todo_tags tt ON t.id = tt.todo_id
WHERE
  t.id = $1
GROUP BY
  t.id;

-- name: ListTodosByWorkspaceID :many
SELECT
  t.id,
  t.title,
  t.status,
  t.created_at,
  t.workspace_id,
  COALESCE(array_remove(array_agg(tt.tag_id), NULL), '{}')::uuid[] AS tags
FROM
  todos t
  LEFT JOIN todo_tags tt ON t.id = tt.todo_id
WHERE
  t.workspace_id = $1
GROUP BY
  t.id
ORDER BY
  t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: AddTagToTodo :exec
INSERT INTO todo_tags(todo_id, tag_id)
  VALUES ($1, $2)
ON CONFLICT
  DO NOTHING;

-- name: RemoveMissingTagsFromTodo :exec
DELETE FROM todo_tags
WHERE todo_id = $1
  AND NOT (tag_id = ANY (sqlc.arg(tags)::uuid[]));

-- name: DeleteTodo :exec
DELETE FROM todos
WHERE id = $1;

