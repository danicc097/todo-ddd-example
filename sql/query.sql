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

-- name: UpdateTodo :exec
UPDATE
  todos
SET
  title = $2,
  completed = $3
WHERE
  id = $1;

