-- name: CreateAccount :one
INSERT INTO accounts (
    id, tenant_id, user_id, name, type, currency, balance_cents, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
) RETURNING id, tenant_id, user_id, name, type, currency, balance_cents, created_at, updated_at, deleted_at;

-- name: GetAccountByID :one
SELECT id, tenant_id, user_id, name, type, currency, balance_cents, created_at, updated_at, deleted_at
FROM accounts
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;

-- name: ListAccountsByTenant :many
SELECT id, tenant_id, user_id, name, type, currency, balance_cents, created_at, updated_at, deleted_at
FROM accounts
WHERE tenant_id = $1 AND deleted_at IS NULL
ORDER BY name ASC;

-- name: UpdateAccount :one
UPDATE accounts
SET name = $3,
    type = $4,
    currency = $5,
    updated_at = NOW()
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
RETURNING id, tenant_id, user_id, name, type, currency, balance_cents, created_at, updated_at, deleted_at;

-- name: UpdateAccountBalance :exec
UPDATE accounts
SET balance_cents = $3,
    updated_at = NOW()
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;

-- name: SoftDeleteAccount :exec
UPDATE accounts
SET deleted_at = NOW()
WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;
