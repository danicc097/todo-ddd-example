-- name: CreateTodo :one
INSERT INTO todos(id, title, status, created_at, updated_at, workspace_id)
  VALUES ($1, $2, $3, $4, $4, $5)
RETURNING
  id, title, status, created_at, workspace_id;

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
  t.created_at DESC;

-- name: UpdateTodo :exec
UPDATE
  todos
SET
  title = $2,
  status = $3,
  updated_at = NOW()
WHERE
  id = $1;

-- name: AddTagToTodo :exec
INSERT INTO todo_tags(todo_id, tag_id)
  VALUES ($1, $2)
  /* we blindly add tags */
ON CONFLICT
  DO NOTHING;

