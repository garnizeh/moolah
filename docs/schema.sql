-- =============================================================================
-- Moolah — PostgreSQL DDL Schema
-- Version: 1.0.0 | Date: 2026-03-06
-- =============================================================================
-- Conventions:
--   • All PKs are CHAR(26) storing ULID strings.
--   • All FKs reference CHAR(26) ULID columns.
--   • Monetary values are INT8 (cents). NEVER use FLOAT for money.
--   • All tenant-scoped tables carry a tenant_id column.
--   • Soft delete via deleted_at TIMESTAMPTZ (NULL = active).
--   • Composite indexes are named: idx_{table}_{columns}.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- Extensions
-- ---------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS "pg_trgm";   -- optional: for fuzzy text search


-- ---------------------------------------------------------------------------
-- ENUM types
-- ---------------------------------------------------------------------------

CREATE TYPE account_type AS ENUM (
    'checking',
    'savings',
    'credit_card',
    'investment'
);

CREATE TYPE transaction_type AS ENUM (
    'income',
    'expense',
    'transfer'
);

CREATE TYPE category_type AS ENUM (
    'income',
    'expense',
    'transfer'
);

CREATE TYPE user_role AS ENUM (
    'sysadmin',
    'admin',
    'member'
);

CREATE TYPE tenant_plan AS ENUM (
    'free',
    'basic',
    'premium'
);

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


-- =============================================================================
-- GLOBAL SCOPE (no tenant isolation)
-- =============================================================================

-- ---------------------------------------------------------------------------
-- tenants — one row per household
-- ---------------------------------------------------------------------------
CREATE TABLE tenants (
    id         CHAR(26)     NOT NULL,
    name       VARCHAR(120) NOT NULL,
    plan       tenant_plan  NOT NULL DEFAULT 'free',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT pk_tenants PRIMARY KEY (id)
);

-- Support global admin lookups (no tenant_id here by design)
CREATE INDEX idx_tenants_deleted_at ON tenants (deleted_at)
    WHERE deleted_at IS NULL;


-- =============================================================================
-- TENANT SCOPE (all tables below require tenant_id = $1 in every query)
-- =============================================================================

-- ---------------------------------------------------------------------------
-- users — household members
-- ---------------------------------------------------------------------------
CREATE TABLE users (
    id             CHAR(26)     NOT NULL,
    tenant_id      CHAR(26)     NOT NULL REFERENCES tenants (id),
    email          VARCHAR(320) NOT NULL,
    name           VARCHAR(120) NOT NULL,
    role           user_role    NOT NULL DEFAULT 'member',
    last_login_at  TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,

    CONSTRAINT pk_users PRIMARY KEY (id),
    -- Email must be unique within a household
    CONSTRAINT uq_users_tenant_email UNIQUE (tenant_id, email)
);

-- Hot path: login & OTP lookup by tenant + email
CREATE INDEX idx_users_tenant_email ON users (tenant_id, email)
    WHERE deleted_at IS NULL;


-- ---------------------------------------------------------------------------
-- otp_requests — time-limited one-time codes
-- ---------------------------------------------------------------------------
CREATE TABLE otp_requests (
    id         CHAR(26)     NOT NULL,
    email      VARCHAR(320) NOT NULL,
    -- Store bcrypt hash of the 6-digit code — never the plain code
    code_hash  TEXT         NOT NULL,
    used       BOOLEAN      NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_otp_requests PRIMARY KEY (id)
);

-- Lookup by email for active (unused, non-expired) codes
CREATE INDEX idx_otp_requests_email_active ON otp_requests (email, expires_at)
    WHERE used = FALSE;


-- ---------------------------------------------------------------------------
-- categories — income/expense categories per household
-- ---------------------------------------------------------------------------
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

-- Category list queries are always scoped to a tenant
CREATE INDEX idx_categories_tenant ON categories (tenant_id)
    WHERE deleted_at IS NULL;


-- ---------------------------------------------------------------------------
-- accounts — financial accounts within a household
-- ---------------------------------------------------------------------------
CREATE TABLE accounts (
    id             CHAR(26)     NOT NULL,
    tenant_id      CHAR(26)     NOT NULL REFERENCES tenants (id),
    -- user_id is the owner; other household members can still view it
    user_id        CHAR(26)     NOT NULL REFERENCES users (id),
    name           VARCHAR(120) NOT NULL,
    type           account_type NOT NULL,
    currency       CHAR(3)      NOT NULL DEFAULT 'BRL',
    -- Cached balance in cents; recomputed on invoice close or reconciliation
    balance_cents  INT8         NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,

    CONSTRAINT pk_accounts PRIMARY KEY (id),
    CONSTRAINT uq_accounts_tenant_name UNIQUE (tenant_id, name)
);

-- Account list queries are always scoped to a tenant
CREATE INDEX idx_accounts_tenant ON accounts (tenant_id)
    WHERE deleted_at IS NULL;

-- Owner-specific queries
CREATE INDEX idx_accounts_tenant_user ON accounts (tenant_id, user_id)
    WHERE deleted_at IS NULL;


-- ---------------------------------------------------------------------------
-- master_purchases — installment purchase header (Phase 2, credit cards)
-- ---------------------------------------------------------------------------
-- One row per installment deal (e.g., "Laptop 12x").
-- Physical transaction rows are only created at invoice-close time.
-- Projections are computed at runtime by the service layer.
-- ---------------------------------------------------------------------------
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

-- Invoice close job queries: find unsettled purchases for an account/date range
CREATE INDEX idx_master_purchases_tenant_account ON master_purchases (tenant_id, account_id, first_due_date)
    WHERE deleted_at IS NULL;


-- ---------------------------------------------------------------------------
-- transactions — individual financial events
-- ---------------------------------------------------------------------------
-- master_purchase_id is set only for installments settled at invoice close.
-- Standalone transactions have master_purchase_id = NULL.
-- ---------------------------------------------------------------------------
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

-- ─── Hot-path indexes for transaction queries ──────────────────────────────

-- Most common query: all transactions for a tenant sorted by date
CREATE INDEX idx_transactions_tenant_date ON transactions (tenant_id, occurred_at DESC)
    WHERE deleted_at IS NULL;

-- Filter by account (account statement view)
CREATE INDEX idx_transactions_tenant_account_date
    ON transactions (tenant_id, account_id, occurred_at DESC)
    WHERE deleted_at IS NULL;

-- Filter by category (budget report)
CREATE INDEX idx_transactions_tenant_category_date
    ON transactions (tenant_id, category_id, occurred_at DESC)
    WHERE deleted_at IS NULL;

-- Installment settlement: find all transactions linked to a master purchase
CREATE INDEX idx_transactions_master_purchase
    ON transactions (master_purchase_id)
    WHERE master_purchase_id IS NOT NULL AND deleted_at IS NULL;


-- ---------------------------------------------------------------------------
-- audit_logs — immutable, append-only record of every state-changing action
-- ---------------------------------------------------------------------------
-- Rows in this table are NEVER soft-deleted or updated.
-- actor_id identifies WHO caused the change:
--   • Regular user  → their ULID (from JWT sub claim)
--   • Sysadmin      → sysadmin ULID
--   • Automated job → the literal string 'SYSTEM'
-- ---------------------------------------------------------------------------
CREATE TABLE audit_logs (
    id          CHAR(26)      NOT NULL,
    -- NULL for global sysadmin operations that span tenants
    tenant_id   CHAR(26)      REFERENCES tenants (id),
    -- ULID of the authenticated user, sysadmin, or 'SYSTEM' for automated jobs
    actor_id    VARCHAR(26)   NOT NULL,
    actor_role  user_role     NOT NULL,
    action      audit_action  NOT NULL,
    entity_type VARCHAR(50)   NOT NULL,  -- 'transaction', 'account', 'user', etc.
    entity_id   CHAR(26),               -- the affected record's ULID
    -- JSON snapshots; old_values is NULL for creates, new_values is NULL for hard deletes
    old_values  JSONB,
    new_values  JSONB,
    ip_address  INET,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_audit_logs PRIMARY KEY (id)
);

-- Tenant-scoped audit trail: who touched what and when
CREATE INDEX idx_audit_logs_tenant_entity
    ON audit_logs (tenant_id, entity_type, entity_id, created_at DESC);

-- Actor history: all actions performed by a specific user/system
CREATE INDEX idx_audit_logs_actor
    ON audit_logs (actor_id, created_at DESC);

-- Chronological audit stream per tenant (compliance reporting)
CREATE INDEX idx_audit_logs_tenant_time
    ON audit_logs (tenant_id, created_at DESC);


-- =============================================================================
-- Utility trigger: keep updated_at current on every UPDATE
-- =============================================================================

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

-- Apply the trigger to all mutable tables
DO $$
DECLARE
    t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY[
        'tenants', 'users', 'categories',
        'accounts', 'master_purchases', 'transactions'
    ]
    LOOP
        EXECUTE format(
            'CREATE TRIGGER trg_%s_updated_at
             BEFORE UPDATE ON %s
             FOR EACH ROW EXECUTE FUNCTION set_updated_at()',
            t, t
        );
    END LOOP;
END;
$$;
