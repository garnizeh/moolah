-- +goose Up
-- +goose StatementBegin
CREATE TABLE transactions (
    id                 CHAR(26)          NOT NULL,
    tenant_id          CHAR(26)          NOT NULL REFERENCES tenants (id),
    account_id         CHAR(26)          NOT NULL REFERENCES accounts (id),
    category_id        CHAR(26)          REFERENCES categories (id),
    user_id            CHAR(26)          NOT NULL REFERENCES users (id),
    -- NULL for non-installment transactions
    master_purchase_id CHAR(26)          REFERENCES master_purchases (id),
    description        VARCHAR(255)      NOT NULL,
    -- Amount in cents (always positive; type determines direction)
    amount_cents       INT8              NOT NULL CHECK (amount_cents > 0),
    type               transaction_type  NOT NULL,
    occurred_at        TIMESTAMPTZ       NOT NULL,
    created_at         TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ,

    CONSTRAINT pk_transactions PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_transactions_tenant_date ON transactions (tenant_id, occurred_at DESC)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_transactions_tenant_account_date
    ON transactions (tenant_id, account_id, occurred_at DESC)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_transactions_tenant_category_date
    ON transactions (tenant_id, category_id, occurred_at DESC)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_transactions_master_purchase
    ON transactions (master_purchase_id)
    WHERE master_purchase_id IS NOT NULL AND deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions;
-- +goose StatementEnd
