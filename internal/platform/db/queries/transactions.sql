-- name: GetTransactionByID :one
SELECT * FROM transactions
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;

-- name: ListTransactionsByTenant :many
SELECT * FROM transactions
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY occurred_at DESC;

-- name: CreateTransaction :one
INSERT INTO transactions (
    id, tenant_id, account_id, category_id, user_id, 
    master_purchase_id, description, amount_cents, type, occurred_at, 
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
) RETURNING *;

-- name: UpdateTransaction :one
UPDATE transactions
SET 
    account_id = $3,
    category_id = $4,
    description = $5,
    amount_cents = $6,
    type = $7,
    occurred_at = $8,
    updated_at = NOW()
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteTransaction :exec
UPDATE transactions
SET deleted_at = NOW()
WHERE tenant_id = $1 AND id = $2;
