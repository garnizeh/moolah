-- name: CreateUser :one
INSERT INTO users (
    id, tenant_id, email, name, role, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, NOW(), NOW()
) RETURNING id, tenant_id, email, name, role, last_login_at, created_at, updated_at, deleted_at;

-- name: GetUserByID :one
SELECT id, tenant_id, email, name, role, last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, tenant_id, email, name, role, last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: ListUsersByTenant :many
SELECT id, tenant_id, email, name, role, last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: UpdateUser :one
UPDATE users
SET email = $3,
    name = $4,
    role = $5,
    updated_at = NOW()
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
RETURNING id, tenant_id, email, name, role, last_login_at, created_at, updated_at, deleted_at;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;
