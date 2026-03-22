-- +goose Up
-- Currency table to handle "Always Cents" and display precision
CREATE TABLE currencies (
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    symbol TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    fallback_decimals INT NOT NULL DEFAULT 2
);

-- Entity table for multiple family members or cost centers
CREATE TABLE entities (
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_entities_deleted ON entities (deleted_at);

-- Account table for financial balances
CREATE TABLE accounts (
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL REFERENCES entities(id),
    currency_id TEXT NOT NULL REFERENCES currencies(id),
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- e.g., 'checking', 'savings', 'credit'
    metadata JSONB NOT NULL DEFAULT '{}',
    balance_cents BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX idx_accounts_entity ON accounts (entity_id);
CREATE INDEX idx_accounts_deleted ON accounts (deleted_at);


-- +goose Down
DROP TABLE accounts;
DROP TABLE entities;
DROP TABLE currencies;
