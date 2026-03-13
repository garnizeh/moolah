-- +goose Up
CREATE TABLE IF NOT EXISTS assets (
    id          VARCHAR(26)  NOT NULL PRIMARY KEY,
    ticker      VARCHAR(20)  NOT NULL,
    isin        VARCHAR(12),
    name        VARCHAR(200) NOT NULL,
    asset_type  asset_type   NOT NULL,
    currency    CHAR(3)      NOT NULL,
    details     TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_assets_ticker UNIQUE (ticker)
);

CREATE TABLE IF NOT EXISTS tenant_asset_configs (
    id          VARCHAR(26)  NOT NULL PRIMARY KEY,
    tenant_id   VARCHAR(26)  NOT NULL REFERENCES tenants(id),
    asset_id    VARCHAR(26)  NOT NULL REFERENCES assets(id),
    name        VARCHAR(200),
    currency    CHAR(3),
    details     TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_tenant_asset_config 
    ON tenant_asset_configs (tenant_id, asset_id) 
    WHERE deleted_at IS NULL;

CREATE INDEX idx_tenant_asset_configs_tenant
    ON tenant_asset_configs(tenant_id, asset_id);
CREATE INDEX idx_tenant_asset_configs_deleted
    ON tenant_asset_configs(tenant_id, deleted_at);

-- +goose Down
DROP TABLE IF EXISTS tenant_asset_configs;
DROP TABLE IF EXISTS assets;
