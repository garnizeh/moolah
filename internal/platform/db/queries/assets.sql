-- name: CreateAsset :one
INSERT INTO assets (id, ticker, isin, name, asset_type, currency, details, created_at)
VALUES (@id, @ticker, @isin, @name, @asset_type, @currency, @details, NOW())
RETURNING *;

-- name: GetAssetByID :one
SELECT * FROM assets
WHERE id = @id;

-- name: GetAssetByTicker :one
SELECT * FROM assets
WHERE ticker = @ticker;

-- name: ListAssets :many
SELECT * FROM assets
ORDER BY ticker ASC;

-- name: DeleteAsset :exec
DELETE FROM assets WHERE id = @id;
