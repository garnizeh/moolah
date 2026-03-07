-- +goose Up
-- +goose StatementBegin
CREATE TABLE tenants (
    id         CHAR(26)     NOT NULL,
    name       VARCHAR(120) NOT NULL,
    plan       tenant_plan  NOT NULL DEFAULT 'free',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_tenants PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_tenants_deleted_at ON tenants (deleted_at)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tenants;
-- +goose StatementEnd
