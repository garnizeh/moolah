-- name: CreatePositionSnapshot :one
INSERT INTO position_snapshots (
    id, tenant_id, position_id, snapshot_date, quantity, 
    avg_cost_cents, last_price_cents, currency
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetPositionSnapshotByID :one
SELECT * FROM position_snapshots
WHERE id = $1 AND tenant_id = $2;

-- name: ListPositionSnapshotsByTenantSince :many
SELECT * FROM position_snapshots
WHERE tenant_id = $1 AND snapshot_date >= $2
ORDER BY snapshot_date DESC, created_at DESC;

-- name: ListPositionSnapshotsByPosition :many
SELECT * FROM position_snapshots
WHERE position_id = $1 AND tenant_id = $2
ORDER BY snapshot_date DESC;
