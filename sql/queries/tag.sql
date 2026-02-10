-- name: CreateTag :one
INSERT INTO tags(id, name)
  VALUES ($1, $2)
RETURNING
  id, name;

-- name: GetTagByID :one
SELECT
  id,
  name
FROM
  tags
WHERE
  id = $1;

-- name: GetTagByName :one
SELECT
  id,
  name
FROM
  tags
WHERE
  name = $1;

