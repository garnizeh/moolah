-- name: AdminListAllAuditLogs :many
SELECT id, tenant_id, actor_id, actor_role, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
FROM audit_logs
WHERE (entity_type = sqlc.narg('entity_type') OR sqlc.narg('entity_type') IS NULL)
  AND (entity_id = sqlc.narg('entity_id') OR sqlc.narg('entity_id') IS NULL)
  AND (actor_id = @actor_id OR @actor_id = '')
  AND (action = sqlc.narg('action') OR sqlc.narg('action') IS NULL)
  AND (created_at >= sqlc.narg('start_date') OR sqlc.narg('start_date') IS NULL)
  AND (created_at <= sqlc.narg('end_date') OR sqlc.narg('end_date') IS NULL)
ORDER BY created_at DESC
LIMIT @limit_off OFFSET @offset_off;
