-- name: CreateTodo :one
INSERT INTO todos(id, title, completed, created_at)
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
  completed = $3
WHERE
  id = $1;
