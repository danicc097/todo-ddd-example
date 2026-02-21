-- name: CreateWorkspace :one
INSERT INTO workspaces(id, name, description, created_at)
  VALUES ($1, $2, $3, $4)
RETURNING
  *;

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
  created_at DESC
LIMIT $1 OFFSET $2;

-- name: DeleteWorkspace :exec
DELETE FROM workspaces
WHERE id = $1;

-- name: UpsertWorkspace :one
INSERT INTO workspaces(id, name, description, created_at)
  VALUES ($1, $2, $3, $4)
ON CONFLICT (id)
  DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description
  RETURNING
    *;

-- name: DeleteWorkspaceMembers :exec
DELETE FROM workspace_members
WHERE workspace_id = $1;

-- name: AddWorkspaceMember :exec
INSERT INTO workspace_members(workspace_id, user_id, role)
  VALUES ($1, $2, $3);

-- name: ListWorkspacesByUserID :many
SELECT
  w.*
FROM
  workspaces w
  JOIN workspace_members wm ON w.id = wm.workspace_id
WHERE
  wm.user_id = $1
ORDER BY
  w.created_at DESC;

-- name: ListWorkspacesWithMembers :many
SELECT
  w.id,
  w.name,
  w.description,
  w.created_at,
  COALESCE(json_agg(json_build_object('user_id', wm.user_id, 'role', wm.role)) FILTER (WHERE wm.user_id IS NOT NULL), '[]')::json AS members
FROM
  workspaces w
  LEFT JOIN workspace_members wm ON w.id = wm.workspace_id
GROUP BY
  w.id
ORDER BY
  w.created_at DESC;

