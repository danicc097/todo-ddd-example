-- name: CreateWorkspace :one
INSERT INTO workspaces(id, name, description, created_at)
  VALUES ($1, $2, $3, $4)
RETURNING
  *;

-- name: AddWorkspaceMember :exec
INSERT INTO workspace_members(workspace_id, user_id, role)
  VALUES ($1, $2, $3);

-- name: GetWorkspaceByID :one
SELECT
  *
FROM
  workspaces
WHERE
  id = $1;

-- name: GetWorkspaceMembers :many
SELECT
  *
FROM
  workspace_members
WHERE
  workspace_id = $1;

-- name: ListWorkspaces :many
SELECT
  *
FROM
  workspaces
ORDER BY
  created_at DESC;

-- name: DeleteWorkspace :exec
DELETE FROM workspaces
WHERE id = $1;

