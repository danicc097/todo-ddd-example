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

