-- +goose Up
-- Currency table to handle "Always Cents" and display precision
CREATE TABLE currencies (
    id TEXT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    symbol TEXT NOT NULL,
    fallback_decimals INT NOT NULL DEFAULT 2,
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Entity table for multiple family members or cost centers
CREATE TABLE entities (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_entities_deleted ON entities (deleted_at);

-- Account table for financial balances
CREATE TABLE accounts (
    id TEXT PRIMARY KEY,
    entity_id TEXT NOT NULL REFERENCES entities(id),
    currency_id TEXT NOT NULL REFERENCES currencies(id),
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- e.g., 'checking', 'savings', 'credit'
    balance_cents BIGINT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_accounts_entity ON accounts (entity_id);
CREATE INDEX idx_accounts_deleted ON accounts (deleted_at);

-- +goose Down
DROP TABLE accounts;
DROP TABLE entities;
DROP TABLE currencies;
