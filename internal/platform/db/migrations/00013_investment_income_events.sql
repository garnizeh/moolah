-- +goose Up
CREATE TABLE IF NOT EXISTS position_snapshots (
    id                VARCHAR(26)    NOT NULL PRIMARY KEY,
    tenant_id         VARCHAR(26)    NOT NULL REFERENCES tenants(id),
    position_id       VARCHAR(26)    NOT NULL REFERENCES positions(id),
    snapshot_date     DATE           NOT NULL,
    quantity          NUMERIC(18,8)  NOT NULL,
    avg_cost_cents    BIGINT         NOT NULL,
    last_price_cents  BIGINT         NOT NULL,
    currency          CHAR(3)        NOT NULL,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_position_snapshot UNIQUE (position_id, snapshot_date)
);

CREATE INDEX idx_position_snapshots_tenant_date ON position_snapshots (tenant_id, snapshot_date);

CREATE TABLE IF NOT EXISTS position_income_events (
    id                VARCHAR(26)       NOT NULL PRIMARY KEY,
    tenant_id         VARCHAR(26)       NOT NULL REFERENCES tenants(id),
    position_id       VARCHAR(26)       NOT NULL REFERENCES positions(id),
    income_type       income_type       NOT NULL,
    amount_cents      BIGINT            NOT NULL CHECK (amount_cents > 0),
    currency          CHAR(3)           NOT NULL,
    event_date        DATE              NOT NULL,
    status            receivable_status NOT NULL DEFAULT 'pending',
    realized_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ       NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_position_income_events_tenant_status ON position_income_events (tenant_id, status);
CREATE INDEX idx_position_income_events_date ON position_income_events (event_date);

-- +goose Down
DROP TABLE IF EXISTS position_income_events;
DROP TABLE IF EXISTS position_snapshots;
