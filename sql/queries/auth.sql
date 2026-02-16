-- name: GetUserAuth :one
SELECT
  *
FROM
  user_auth
WHERE
  user_id = $1;

-- name: UpsertUserAuth :exec
INSERT INTO user_auth(user_id, totp_status, totp_secret_cipher, totp_secret_nonce, password_hash)
  VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id)
  DO UPDATE SET
    totp_status = EXCLUDED.totp_status,
    totp_secret_cipher = EXCLUDED.totp_secret_cipher,
    totp_secret_nonce = EXCLUDED.totp_secret_nonce,
    password_hash = EXCLUDED.password_hash;

