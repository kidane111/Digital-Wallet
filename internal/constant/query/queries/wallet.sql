-- name: CreateWallet :one
INSERT INTO wallets (user_id, currency)
VALUES ($1, $2)
RETURNING *;

-- name: GetWallet :one
SELECT * FROM wallets WHERE id = $1 LIMIT 1;

-- name: GetWalletByUser :one
SELECT * FROM wallets WHERE user_id = $1 AND currency = $2 LIMIT 1;

-- name: UpdateWalletBalance :exec
UPDATE wallets
SET 
    balance = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: ListWallets :many
SELECT * FROM wallets
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: RegisterBill :one
INSERT INTO bills (
    id,
    user_id,
    provider,
    account_number,
    amount,
    fee,
    currency,
    due_date,
    status,
    transaction_id,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetBillByReferenceID :one
SELECT * FROM bills
WHERE id = $1;