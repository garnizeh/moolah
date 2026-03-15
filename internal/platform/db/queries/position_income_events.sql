-- name: CreatePositionIncomeEvent :one
INSERT INTO position_income_events (
    id, tenant_id, position_id, income_type, amount_cents, 
    currency, event_date, status, realized_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetPositionIncomeEventByID :one
SELECT * FROM position_income_events
WHERE id = $1 AND tenant_id = $2;

-- name: ListPositionIncomeEventsByTenant :many
SELECT * FROM position_income_events
WHERE tenant_id = $1
ORDER BY event_date DESC;

-- name: ListPendingIncomeEvents :many
SELECT * FROM position_income_events
WHERE tenant_id = $1 AND status = 'pending'
ORDER BY event_date ASC;

-- name: UpdateIncomeEventStatus :one
UPDATE position_income_events
SET status = $3,
    realized_at = $4,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING *;
