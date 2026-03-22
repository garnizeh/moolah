-- name: CreateCurrency :one
INSERT INTO currencies (
    id, code, symbol, config, fallback_decimals
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;


-- name: GetCurrency :one
SELECT * FROM currencies
WHERE id = $1 LIMIT 1;

-- name: ListCurrencies :many
SELECT * FROM currencies
ORDER BY code;

-- name: UpdateCurrency :one
UPDATE currencies
SET 
    code = $2,
    symbol = $3,
    config = $4,
    fallback_decimals = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

