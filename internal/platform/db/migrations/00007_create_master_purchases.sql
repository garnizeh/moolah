-- +goose Up
-- +goose StatementBegin
CREATE TABLE master_purchases (
    id                          CHAR(26)     NOT NULL,
    tenant_id                   CHAR(26)     NOT NULL REFERENCES tenants (id),
    account_id                  CHAR(26)     NOT NULL REFERENCES accounts (id),
    category_id                 CHAR(26)     REFERENCES categories (id),
    user_id                     CHAR(26)     NOT NULL REFERENCES users (id),
    description                 VARCHAR(255) NOT NULL,
    total_amount_cents          INT8         NOT NULL CHECK (total_amount_cents > 0),
    installment_count           SMALLINT     NOT NULL CHECK (installment_count > 0),
    -- Per-installment amount (integer division; remainder absorbed by last installment)
    installment_cents           INT8         NOT NULL CHECK (installment_cents > 0),
    first_due_date              DATE         NOT NULL,
    -- Tracks which installments have been settled into physical transactions
    last_settled_installment    SMALLINT     NOT NULL DEFAULT 0,
    created_at                  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,

    CONSTRAINT pk_master_purchases PRIMARY KEY (id),
    CONSTRAINT chk_master_purchases_settled
        CHECK (last_settled_installment <= installment_count)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_master_purchases_tenant_account ON master_purchases (tenant_id, account_id, first_due_date)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS master_purchases;
-- +goose StatementEnd
