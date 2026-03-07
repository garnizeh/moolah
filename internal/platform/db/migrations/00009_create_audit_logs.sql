-- +goose Up
-- +goose StatementBegin
CREATE TABLE audit_logs (
    id          CHAR(26)      NOT NULL,
    -- NULL for global sysadmin operations that span tenants
    tenant_id   CHAR(26)      REFERENCES tenants (id),
    -- ULID of the authenticated user, sysadmin, or 'SYSTEM' for automated jobs
    actor_id    VARCHAR(26)   NOT NULL,
    actor_role  user_role     NOT NULL,
    action      audit_action  NOT NULL,
    entity_type VARCHAR(50)   NOT NULL,  -- 'transaction', 'account', 'user', etc.
    entity_id   CHAR(26),               -- the affected record's ULID
    -- JSON snapshots; old_values is NULL for creates, new_values is NULL for hard deletes
    old_values  JSONB,
    new_values  JSONB,
    ip_address  INET,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_audit_logs PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_audit_logs_tenant_entity
    ON audit_logs (tenant_id, entity_type, entity_id, created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_audit_logs_actor
    ON audit_logs (actor_id, created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_audit_logs_tenant_time
    ON audit_logs (tenant_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_logs;
-- +goose StatementEnd
