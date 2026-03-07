-- name: AdminListAllUsers :many
SELECT id, tenant_id, email, name, role, last_login_at, created_at, updated_at, deleted_at
FROM users
ORDER BY created_at ASC;

-- name: AdminGetUserByID :one
SELECT id, tenant_id, email, name, role, last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE id = $1;

-- name: AdminForceDeleteUser :exec
DELETE FROM users
WHERE id = $1;
