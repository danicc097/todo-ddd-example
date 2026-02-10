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
LIMIT 100
/* lock per tx in replica: e.g. 200 rows - a locks 100, b locks next 100, ... */
FOR UPDATE
  SKIP LOCKED;

-- name: MarkOutboxEventProcessed :exec
UPDATE
  outbox
SET
  processed_at = NOW()
WHERE
  id = $1;

