-- +goose Up
-- +goose StatementBegin
CREATE TYPE account_type AS ENUM (
    'checking',
    'savings',
    'credit_card',
    'investment'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE transaction_type AS ENUM (
    'income',
    'expense',
    'transfer'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE category_type AS ENUM (
    'income',
    'expense',
    'transfer'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE user_role AS ENUM (
    'sysadmin',
    'admin',
    'member'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE tenant_plan AS ENUM (
    'free',
    'basic',
    'premium'
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TYPE audit_action AS ENUM (
    'create',
    'update',
    'soft_delete',
    'restore',
    'login',
    'login_failed',
    'otp_requested',
    'otp_verified'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS audit_action;
DROP TYPE IF EXISTS tenant_plan;
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS category_type;
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS account_type;
-- +goose StatementEnd
