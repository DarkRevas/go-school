-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token, expires_at, revoked, created_at;

-- name: GetRefreshToken :one
SELECT id, user_id, token, expires_at, revoked, created_at
FROM refresh_tokens
WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1;

-- name: GetActiveRefreshToken :one
SELECT id, user_id, token, expires_at, revoked, created_at
FROM refresh_tokens
WHERE token = $1 AND revoked = FALSE AND expires_at > NOW();
