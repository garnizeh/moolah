-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id             CHAR(26)     NOT NULL,
    tenant_id      CHAR(26)     NOT NULL REFERENCES tenants (id),
    email          VARCHAR(320) NOT NULL,
    name           VARCHAR(120) NOT NULL,
    role           user_role    NOT NULL DEFAULT 'member',
    last_login_at  TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,

    CONSTRAINT pk_users PRIMARY KEY (id),
    -- Email must be unique within a household
    CONSTRAINT uq_users_tenant_email UNIQUE (tenant_id, email)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_users_tenant_email ON users (tenant_id, email)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
