-- name: CreateMasterPurchase :one
INSERT INTO master_purchases (
    id,
    tenant_id,
    account_id,
    category_id,
    user_id,
    description,
    total_amount_cents,
    installment_count,
    closing_day,
    first_installment_date
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetMasterPurchaseByID :one
SELECT * FROM master_purchases
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: ListMasterPurchasesByTenant :many
SELECT * FROM master_purchases
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListMasterPurchasesByAccount :many
SELECT * FROM master_purchases
WHERE tenant_id = $1 AND account_id = $2 AND deleted_at IS NULL
ORDER BY first_installment_date DESC;

-- name: ListPendingMasterPurchasesByClosingDay :many
SELECT * FROM master_purchases
WHERE tenant_id = $1 
  AND status = 'open' 
  AND closing_day = $2
  AND deleted_at IS NULL;

-- name: UpdateMasterPurchase :one
UPDATE master_purchases
SET 
    category_id = COALESCE(NULLIF(sqlc.arg('category_id'), ''), category_id),
    description = COALESCE(sqlc.narg('description'), description),
    status = COALESCE(sqlc.narg('status'), status),
    paid_installments = COALESCE(sqlc.narg('paid_installments'), paid_installments),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: IncrementPaidInstallments :one
UPDATE master_purchases
SET 
    paid_installments = paid_installments + 1,
    status = CASE 
        WHEN paid_installments + 1 >= installment_count THEN 'closed'::master_purchase_status 
        ELSE status 
    END,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteMasterPurchase :exec
UPDATE master_purchases
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;
