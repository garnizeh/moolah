-- name: CreateTransaction :one
INSERT INTO transactions (
    id, tenant_id, account_id, category_id, user_id, 
    master_purchase_id, description, amount_cents, type, 
    occurred_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
) RETURNING id, tenant_id, account_id, category_id, user_id, master_purchase_id, description, amount_cents, type, occurred_at, created_at, updated_at, deleted_at;

-- name: GetTransactionByID :one
SELECT id, tenant_id, account_id, category_id, user_id, master_purchase_id, description, amount_cents, type, occurred_at, created_at, updated_at, deleted_at
FROM transactions
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;

-- name: ListTransactionsByTenant :many
-- One query with optional filters using COALESCE or IS NULL pattern
SELECT id, tenant_id, account_id, category_id, user_id, master_purchase_id, description, amount_cents, type, occurred_at, created_at, updated_at, deleted_at
FROM transactions
WHERE tenant_id = $1 
    AND (sqlc.narg('account_id')::CHAR(26) IS NULL OR account_id = sqlc.narg('account_id'))
    AND (sqlc.narg('category_id')::CHAR(26) IS NULL OR category_id = sqlc.narg('category_id'))
    AND (sqlc.narg('start_date')::TIMESTAMPTZ IS NULL OR occurred_at >= sqlc.narg('start_date'))
    AND (sqlc.narg('end_date')::TIMESTAMPTZ IS NULL OR occurred_at <= sqlc.narg('end_date'))
    AND deleted_at IS NULL
ORDER BY occurred_at DESC
LIMIT @limit_off OFFSET @offset_off;

-- name: UpdateTransaction :one
UPDATE transactions
SET account_id = $3,
    category_id = $4,
    description = $5,
    amount_cents = $6,
    type = $7,
    occurred_at = $8,
    master_purchase_id = $9,
    updated_at = NOW()
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
RETURNING id, tenant_id, account_id, category_id, user_id, master_purchase_id, description, amount_cents, type, occurred_at, created_at, updated_at, deleted_at;

-- name: SoftDeleteTransaction :exec
UPDATE transactions
SET deleted_at = NOW()
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;
