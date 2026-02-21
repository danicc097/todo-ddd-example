-- name: GetIdempotencyKey :one
SELECT
  *
FROM
  idempotency_keys
WHERE
  id = $1;

-- name: TryLockIdempotencyKey :execrows
INSERT INTO idempotency_keys(id, response_status, response_headers, response_body, locked_at)
  VALUES ($1, 0, '{}', '', NOW())
ON CONFLICT (id)
  DO UPDATE SET
    locked_at = NOW()
  WHERE
    idempotency_keys.response_status = 0
    AND idempotency_keys.locked_at < NOW() - INTERVAL '1 minute';

-- name: UpdateIdempotencyKey :exec
UPDATE
  idempotency_keys
SET
  response_status = $2,
  response_headers = $3,
  response_body = $4
WHERE
  id = $1;

-- name: DeleteIdempotencyKey :exec
DELETE FROM idempotency_keys
WHERE id = $1;

