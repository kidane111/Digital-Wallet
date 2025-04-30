-- name: FindUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: FindUserByPhone :one
SELECT *
FROM users
WHERE phone = $1;

-- name: UpdateLastLogin :exec
UPDATE users
SET last_login = NOW()
WHERE id = $1;

-- name: UserRegister :one
INSERT INTO users (
    name,
    email,
    phone,
    password,
    tier,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING id, name, email, phone, tier, status, created_at, updated_at;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateBillStatus :exec
UPDATE bills 
SET status = $2, 
    updated_at = NOW()
WHERE id = $1;

-- name: GetBillByID :one
SELECT * FROM bills WHERE id = $1;

-- name: GetBalanceByUserID :one
SELECT balance
FROM wallets
WHERE user_id = $1;


-- name: GetTierConfig :one
SELECT * FROM tier_configs WHERE tier = $1;

-- name: UpdateTierConfig :exec
UPDATE tier_configs
SET 
    name = COALESCE($2, name),
    monthly_fee = COALESCE($3, monthly_fee),
    max_transactions = COALESCE($4, max_transactions),
    max_amount = COALESCE($5, max_amount),
    features = COALESCE($6, features),
    updated_at = NOW()
WHERE tier = $1;


