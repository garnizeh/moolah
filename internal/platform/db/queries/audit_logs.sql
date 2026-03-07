-- name: CreateAuditLog :one
INSERT INTO audit_logs (
    id, tenant_id, actor_id, actor_role, action, 
    entity_type, entity_id, old_values, new_values, 
    ip_address, user_agent, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW()
) RETURNING id, tenant_id, actor_id, actor_role, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at;

-- name: ListAuditLogsByTenant :many
SELECT id, tenant_id, actor_id, actor_role, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
FROM audit_logs
WHERE (tenant_id = $1 OR tenant_id IS NULL) -- Allow global sysadmin logs
ORDER BY created_at DESC
LIMIT @limit_off OFFSET @offset_off;

-- name: ListAuditLogsByEntity :many
SELECT id, tenant_id, actor_id, actor_role, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
FROM audit_logs
WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
ORDER BY created_at DESC;
