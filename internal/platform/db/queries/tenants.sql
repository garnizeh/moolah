-- name: CreateTenant :one
INSERT INTO tenants (
    id, name, plan, created_at, updated_at
) VALUES (
    $1, $2, $3, NOW(), NOW()
) RETURNING id, name, plan, created_at, updated_at, deleted_at;

-- name: GetTenantByID :one
SELECT id, name, plan, created_at, updated_at, deleted_at
FROM tenants
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateTenant :one
UPDATE tenants
SET name = $2,
    plan = $3,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, name, plan, created_at, updated_at, deleted_at;

-- name: SoftDeleteTenant :exec
UPDATE tenants
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
