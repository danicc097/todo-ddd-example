-- name: SaveOutboxEvent :exec
INSERT INTO outbox(id, event_type, aggregate_type, aggregate_id, payload, headers)
  VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetUnprocessedOutboxEvents :many
SELECT
  *
FROM
  outbox
WHERE
  processed_at IS NULL
  -- max backoff 1024s
  AND (last_attempted_at IS NULL
    OR NOW() >= last_attempted_at + make_interval(secs := power(2, LEAST(retries, 10))::int))
ORDER BY
  created_at ASC
LIMIT 10
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

-- name: UpdateOutboxRetries :exec
UPDATE
  outbox
SET
  retries = retries + 1,
  last_error = $2,
  last_attempted_at = NOW()
WHERE
  id = $1;

-- name: GetOutboxLag :one
SELECT
  COUNT(*) AS total_lag,
  COALESCE(EXTRACT(EPOCH FROM (NOW() - MIN(created_at))), 0)::float AS max_age_seconds
FROM
  outbox
WHERE
  processed_at IS NULL;

-- name: DeleteProcessedOutboxEvents :exec
DELETE FROM outbox
WHERE processed_at < NOW() - INTERVAL '7 days';

