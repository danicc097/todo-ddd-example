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

-- name: GetUserByEmail :one
SELECT
  *
FROM
  users
WHERE
  email = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

