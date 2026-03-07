-- name: GetCategoryByID :one
SELECT * FROM categories
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;

-- name: ListCategoriesByTenant :many
SELECT * FROM categories
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: ListChildCategories :many
SELECT * FROM categories
WHERE tenant_id = $1 AND parent_id = $2 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: CreateCategory :one
INSERT INTO categories (
    id, tenant_id, parent_id, name, icon, color, type, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
) RETURNING *;

-- name: UpdateCategory :one
UPDATE categories
SET 
    parent_id = $3,
    name = $4,
    icon = $5,
    color = $6,
    type = $7,
    updated_at = NOW()
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCategory :exec
UPDATE categories
SET deleted_at = NOW()
WHERE tenant_id = $1 AND id = $2;
