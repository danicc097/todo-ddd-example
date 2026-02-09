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
