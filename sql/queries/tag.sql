-- name: CreateTag :one
INSERT INTO tags(id, name, workspace_id)
  VALUES ($1, $2, $3)
RETURNING
  *;

-- name: GetTagByID :one
SELECT
  *
FROM
  tags
WHERE
  id = $1;

-- name: GetTagByName :one
SELECT
  *
FROM
  tags
WHERE
  workspace_id = $1
  AND name = $2;

-- name: ListTagsByWorkspaceID :many
SELECT
  *
FROM
  tags
WHERE
  workspace_id = $1
ORDER BY
  name ASC;

-- name: DeleteTag :exec
DELETE FROM tags
WHERE id = $1;

