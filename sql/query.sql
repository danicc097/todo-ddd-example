-- name: CreateTodo :one
INSERT INTO todos(id, title, status, created_at)
  VALUES ($1, $2, $3, $4)
RETURNING
  *;

-- name: GetTodoByID :one
SELECT
  *
FROM
  todos
WHERE
  id = $1;

-- name: ListTodos :many
SELECT
  *
FROM
  todos
ORDER BY
  created_at DESC;

-- name: UpdateTodo :exec
UPDATE
  todos
SET
  title = $2,
  status = $3
WHERE
  id = $1;

-- name: CreateUser :one
INSERT INTO users(id, email, name, created_at)
  VALUES ($1, $2, $3, $4)
RETURNING
  *;

-- name: GetUserByID :one
SELECT
  *
FROM
  users
WHERE
  id = $1;

-- name: AddTagToTodo :exec
INSERT INTO todo_tags(todo_id, tag_id)
  VALUES ($1, $2);

-- name: SaveOutboxEvent :exec
INSERT INTO outbox(id, event_type, payload)
  VALUES ($1, $2, $3);

-- name: GetUnprocessedOutboxEvents :many
SELECT
  *
FROM
  outbox
WHERE
  processed_at IS NULL
ORDER BY
  created_at ASC
LIMIT 100;

-- name: MarkOutboxEventProcessed :exec
UPDATE
  outbox
SET
  processed_at = NOW()
WHERE
  id = $1;

