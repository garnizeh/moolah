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
    sqlc.arg('id'),
    sqlc.arg('tenant_id'),
    sqlc.narg('parent_id')::CHAR(26),
    sqlc.arg('name'),
    sqlc.arg('icon'),
    sqlc.arg('color'),
    sqlc.arg('type'),
    NOW(),
    NOW()
) RETURNING *;

-- name: UpdateCategory :one
UPDATE categories
SET 
    parent_id = sqlc.narg('parent_id')::CHAR(26),
    name = sqlc.arg('name'),
    icon = sqlc.arg('icon'),
    color = sqlc.arg('color'),
    type = sqlc.arg('type'),
    updated_at = NOW()
WHERE tenant_id = sqlc.arg('tenant_id') AND id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCategory :exec
UPDATE categories
SET deleted_at = NOW()
WHERE tenant_id = $1 AND id = $2;
