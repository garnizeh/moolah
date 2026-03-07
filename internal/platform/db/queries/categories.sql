-- name: CreateCategory :one
INSERT INTO categories (
    id, tenant_id, parent_id, name, icon, color, type, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
) RETURNING id, tenant_id, parent_id, name, icon, color, type, created_at, updated_at, deleted_at;

-- name: GetCategoryByID :one
SELECT id, tenant_id, parent_id, name, icon, color, type, created_at, updated_at, deleted_at
FROM categories
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;

-- name: ListCategoriesByTenant :many
SELECT id, tenant_id, parent_id, name, icon, color, type, created_at, updated_at, deleted_at
FROM categories
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: ListRootCategoriesByTenant :many
SELECT id, tenant_id, parent_id, name, icon, color, type, created_at, updated_at, deleted_at
FROM categories
WHERE tenant_id = $1 AND parent_id IS NULL AND deleted_at IS NULL
ORDER BY name ASC;

-- name: ListChildCategories :many
SELECT id, tenant_id, parent_id, name, icon, color, type, created_at, updated_at, deleted_at
FROM categories
WHERE tenant_id = $1 AND parent_id = $2 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: UpdateCategory :one
UPDATE categories
SET name = $3,
    icon = $4,
    color = $5,
    type = $6,
    parent_id = $7,
    updated_at = NOW()
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
RETURNING id, tenant_id, parent_id, name, icon, color, type, created_at, updated_at, deleted_at;

-- name: SoftDeleteCategory :exec
UPDATE categories
SET deleted_at = NOW()
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;
