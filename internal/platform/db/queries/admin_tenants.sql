-- name: AdminListAllTenants :many
SELECT id, name, plan, created_at, updated_at, deleted_at
FROM tenants
WHERE (@with_deleted::boolean = true OR deleted_at IS NULL)
ORDER BY name ASC;

-- name: AdminGetTenantByID :one
SELECT id, name, plan, created_at, updated_at, deleted_at
FROM tenants
WHERE id = $1;

-- name: AdminUpdateTenantPlan :one
UPDATE tenants
SET plan = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, plan, created_at, updated_at, deleted_at;

-- name: AdminSuspendTenant :exec
UPDATE tenants
SET deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: AdminRestoreTenant :exec
UPDATE tenants
SET deleted_at = NULL,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NOT NULL;

-- name: AdminHardDeleteTenant :exec
DELETE FROM tenants
WHERE id = $1;
