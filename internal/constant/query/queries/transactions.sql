-- name: CreateTransaction :one
INSERT INTO transactions (
    wallet_id,
    amount,
    transaction_type,
    description,
    reference,
    status,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetTransaction :one
SELECT * FROM transactions WHERE id = $1 LIMIT 1;

-- name: GetTransactionByReference :one
SELECT * FROM transactions WHERE reference = $1 LIMIT 1;

-- name: ListWalletTransactions :many
SELECT * FROM transactions
WHERE wallet_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTransactionsByType :many
SELECT * FROM transactions
WHERE wallet_id = $1 AND transaction_type = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateTransactionStatus :exec
UPDATE transactions
SET 
    status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: SumTransactionsByType :one
SELECT COALESCE(SUM(amount), 0) FROM transactions
WHERE wallet_id = $1 AND transaction_type = $2 AND status = 'completed';


