-- name: CreatePosition :one
INSERT INTO positions (
    id, tenant_id, asset_id, account_id, quantity, avg_cost_cents, 
    last_price_cents, currency, purchased_at, income_type, 
    income_interval_days, income_amount_cents, income_rate_bps, 
    next_income_at, maturity_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: GetPositionByID :one
SELECT * FROM positions
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: ListPositionsByTenant :many
SELECT * FROM positions
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListPositionsByAccount :many
SELECT * FROM positions
WHERE tenant_id = $1 AND account_id = $2 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListPositionsDueIncome :many
-- Scheduler processes all tenants globally.
SELECT * FROM positions
WHERE income_type != 'none' 
  AND next_income_at <= $1 
  AND deleted_at IS NULL;

-- name: UpdatePosition :one
UPDATE positions
SET quantity = $3,
    avg_cost_cents = $4,
    last_price_cents = $5,
    income_type = $6,
    income_interval_days = $7,
    income_amount_cents = $8,
    income_rate_bps = $9,
    next_income_at = $10,
    maturity_at = $11,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: UpdatePositionNextIncome :exec
UPDATE positions
SET next_income_at = $3,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;

-- name: SoftDeletePosition :exec
UPDATE positions
SET deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2;
