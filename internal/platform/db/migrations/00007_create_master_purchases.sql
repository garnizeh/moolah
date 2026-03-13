-- +goose Up
-- +goose StatementBegin
CREATE TYPE master_purchase_status AS ENUM ('open', 'closed');

CREATE TABLE master_purchases (
    id                    VARCHAR(26)              NOT NULL,
    tenant_id             VARCHAR(26)              NOT NULL,
    account_id            VARCHAR(26)              NOT NULL,
    category_id           VARCHAR(26),
    user_id               VARCHAR(26)              NOT NULL,
    description           VARCHAR(255)             NOT NULL,
    status                master_purchase_status   NOT NULL DEFAULT 'open',
    total_amount_cents    BIGINT                   NOT NULL CHECK (total_amount_cents > 0),
    installment_count     SMALLINT                 NOT NULL CHECK (installment_count BETWEEN 2 AND 48),
    paid_installments     SMALLINT                 NOT NULL DEFAULT 0 CHECK (paid_installments >= 0),
    closing_day           SMALLINT                 NOT NULL CHECK (closing_day BETWEEN 1 AND 28),
    first_installment_date DATE                    NOT NULL,
    created_at            TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ,

    CONSTRAINT pk_master_purchases           PRIMARY KEY (id),
    CONSTRAINT fk_mp_tenant                  FOREIGN KEY (tenant_id)   REFERENCES tenants(id),
    CONSTRAINT fk_mp_account                 FOREIGN KEY (account_id)  REFERENCES accounts(id),
    CONSTRAINT fk_mp_category                FOREIGN KEY (category_id) REFERENCES categories(id),
    CONSTRAINT fk_mp_user                    FOREIGN KEY (user_id)     REFERENCES users(id),
    CONSTRAINT chk_paid_lte_total            CHECK (paid_installments <= installment_count)
);

CREATE INDEX idx_mp_tenant_account ON master_purchases (tenant_id, account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_mp_pending_close  ON master_purchases (tenant_id, status, closing_day) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS master_purchases;
DROP TYPE IF EXISTS master_purchase_status;
-- +goose StatementEnd
