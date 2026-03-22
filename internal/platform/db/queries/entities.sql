-- name: CreateEntity :one
INSERT INTO entities (
    id, name, role, metadata
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetEntity :one
SELECT * FROM entities
WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: ListEntities :many
SELECT * FROM entities
WHERE deleted_at IS NULL
ORDER BY name;

-- name: UpdateEntity :one
UPDATE entities
SET 
    name = $2,
    role = $3,
    metadata = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteEntity :exec
UPDATE entities
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1;
