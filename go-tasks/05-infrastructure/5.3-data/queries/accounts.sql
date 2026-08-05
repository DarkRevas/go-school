-- name: CreateAccount :one
INSERT INTO accounts (user_id, balance) VALUES ($1, $2) RETURNING id, user_id, balance, created_at;

-- name: GetAccount :one
SELECT id, user_id, balance, created_at FROM accounts WHERE id = $1;

-- name: GetAccountByUserID :one
SELECT id, user_id, balance, created_at FROM accounts WHERE user_id = $1;

-- name: UpdateAccountBalance :one
UPDATE accounts SET balance = $2 WHERE id = $1 RETURNING id, user_id, balance, created_at;

-- name: Transfer :exec
-- Transfer is handled in a transaction with two updates
-- This query is a placeholder; actual transfer logic is in code
