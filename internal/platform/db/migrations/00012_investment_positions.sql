-- +goose Up
CREATE TABLE IF NOT EXISTS positions (
    id                    VARCHAR(26)    NOT NULL PRIMARY KEY,
    tenant_id             VARCHAR(26)    NOT NULL REFERENCES tenants(id),
    asset_id              VARCHAR(26)    NOT NULL REFERENCES assets(id),
    account_id            VARCHAR(26)    NOT NULL REFERENCES accounts(id),
    quantity              NUMERIC(18,8)  NOT NULL CHECK (quantity >= 0),
    avg_cost_cents        BIGINT         NOT NULL DEFAULT 0 CHECK (avg_cost_cents >= 0),
    last_price_cents      BIGINT         NOT NULL DEFAULT 0 CHECK (last_price_cents >= 0),
    currency              CHAR(3)        NOT NULL,
    purchased_at          TIMESTAMPTZ    NOT NULL,
    income_type           income_type    NOT NULL DEFAULT 'none',
    income_interval_days  INT            CHECK (income_interval_days > 0),
    income_amount_cents   BIGINT         CHECK (income_amount_cents >= 0),
    income_rate_bps       INT            CHECK (income_rate_bps >= 0),
    next_income_at        TIMESTAMPTZ,
    maturity_at           TIMESTAMPTZ,
    created_at            TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_position 
    ON positions (tenant_id, asset_id, account_id) 
    WHERE deleted_at IS NULL;

CREATE INDEX idx_positions_tenant_asset ON positions (tenant_id, asset_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_positions_tenant_account ON positions (tenant_id, account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_positions_next_income ON positions (next_income_at) WHERE deleted_at IS NULL AND next_income_at IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS positions;
