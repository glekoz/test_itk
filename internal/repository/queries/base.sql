-- name: CreateWallet :exec
INSERT INTO wallets (id, amount)
VALUES ($1, 0);

-- name: GetBalance :one
SELECT amount 
FROM wallets
WHERE id = $1;

-- name: Deposit :execrows
UPDATE wallets
SET amount = amount + $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: Withdraw :execrows
UPDATE wallets
SET amount = amount - $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CreateTransaction :exec
INSERT INTO transactions (id, wallet_id, amount, operation_type)
VALUES ($1, $2, $3, $4);