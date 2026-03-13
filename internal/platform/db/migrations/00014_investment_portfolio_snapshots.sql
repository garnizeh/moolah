-- +goose Up
CREATE TABLE IF NOT EXISTS portfolio_snapshots (
    id                 VARCHAR(26)  NOT NULL PRIMARY KEY,
    tenant_id          VARCHAR(26)  NOT NULL REFERENCES tenants(id),
    snapshot_date      DATE         NOT NULL,
    total_value_cents  BIGINT       NOT NULL DEFAULT 0,
    total_income_cents BIGINT       NOT NULL DEFAULT 0,
    currency           CHAR(3)      NOT NULL,
    details            JSONB,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_portfolio_snapshot_tenant_date UNIQUE (tenant_id, snapshot_date)
);

-- +goose Down
DROP TABLE IF EXISTS portfolio_snapshots;
