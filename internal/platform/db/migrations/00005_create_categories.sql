-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories (
    id        CHAR(26)      NOT NULL,
    tenant_id CHAR(26)      NOT NULL REFERENCES tenants (id),
    -- parent_id allows nested categories (e.g. Food > Restaurants)
    parent_id CHAR(26)      REFERENCES categories (id),
    name      VARCHAR(80)   NOT NULL,
    icon      VARCHAR(40),
    color     CHAR(7),       -- hex color, e.g. #FF5733
    type      category_type NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_categories PRIMARY KEY (id),
    CONSTRAINT uq_categories_tenant_name UNIQUE (tenant_id, name)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_categories_tenant ON categories (tenant_id)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS categories;
-- +goose StatementEnd
