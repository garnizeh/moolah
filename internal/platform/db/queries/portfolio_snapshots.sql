-- name: CreatePortfolioSnapshot :one
INSERT INTO portfolio_snapshots (
    id, tenant_id, snapshot_date, total_value_cents, 
    total_income_cents, currency, details
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetPortfolioSnapshotByDate :one
SELECT * FROM portfolio_snapshots
WHERE tenant_id = $1 AND snapshot_date = $2;

-- name: ListPortfolioSnapshots :many
SELECT * FROM portfolio_snapshots
WHERE tenant_id = $1
ORDER BY snapshot_date DESC;
