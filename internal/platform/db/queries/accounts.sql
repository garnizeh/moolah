-- name: CreateAccount :one
INSERT INTO accounts (
    id, entity_id, currency_id, name, type, balance_cents, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: ListAccountsByEntity :many
SELECT * FROM accounts
WHERE entity_id = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: UpdateAccountBalance :one
UPDATE accounts
SET 
    balance_cents = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteAccount :exec
UPDATE accounts
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

