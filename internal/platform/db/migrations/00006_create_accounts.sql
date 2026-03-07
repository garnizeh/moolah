-- +goose Up
-- +goose StatementBegin
CREATE TABLE accounts (
    id             CHAR(26)     NOT NULL,
    tenant_id      CHAR(26)     NOT NULL REFERENCES tenants (id),
    -- user_id is the owner; other household members can still view it
    user_id        CHAR(26)     NOT NULL REFERENCES users (id),
    name           VARCHAR(120) NOT NULL,
    type           account_type NOT NULL,
    currency       CHAR(3)      NOT NULL DEFAULT 'BRL',
    -- Cached balance in cents; recomputed on invoice close or reconciliation
    balance_cents  INT8         NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,

    CONSTRAINT pk_accounts PRIMARY KEY (id),
    CONSTRAINT uq_accounts_tenant_name UNIQUE (tenant_id, name)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_accounts_tenant ON accounts (tenant_id)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_accounts_tenant_user ON accounts (tenant_id, user_id)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS accounts;
-- +goose StatementEnd
