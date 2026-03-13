-- name: UpsertTenantAssetConfig :one
INSERT INTO tenant_asset_configs (id, tenant_id, asset_id, name, currency, details, created_at, updated_at)
VALUES (@id, @tenant_id, @asset_id, @name, @currency, @details, NOW(), NOW())
ON CONFLICT (tenant_id, asset_id) WHERE deleted_at IS NULL
DO UPDATE SET
    name       = EXCLUDED.name,
    currency   = EXCLUDED.currency,
    details    = EXCLUDED.details,
    updated_at = NOW()
RETURNING *;

-- name: GetTenantAssetConfigByAssetID :one
SELECT * FROM tenant_asset_configs
WHERE tenant_id = @tenant_id
  AND asset_id  = @asset_id
  AND deleted_at IS NULL;

-- name: ListTenantAssetConfigs :many
SELECT * FROM tenant_asset_configs
WHERE tenant_id = @tenant_id
  AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: SoftDeleteTenantAssetConfig :exec
UPDATE tenant_asset_configs
SET deleted_at = NOW()
WHERE tenant_id = @tenant_id
  AND asset_id  = @asset_id
  AND deleted_at IS NULL;

-- name: GetAssetWithTenantConfig :one
-- Returns the global asset with tenant overrides applied via COALESCE.
-- Use this query in all asset-display contexts (ADR §2.7).
SELECT
    a.id,
    a.ticker,
    a.isin,
    a.asset_type,
    a.created_at,
    COALESCE(tac.name,     a.name)     AS name,
    COALESCE(tac.currency, a.currency) AS currency,
    COALESCE(tac.details,  a.details)  AS details,
    tac.id         AS config_id,
    tac.created_at AS tenant_added_at
FROM assets a
LEFT JOIN tenant_asset_configs tac
       ON tac.asset_id = a.id
      AND tac.tenant_id = @tenant_id
      AND tac.deleted_at IS NULL
WHERE a.id = @asset_id;
