# Task 3.2 — Goose Migrations: All Phase 3 Tables (6 tables + 3 enums)

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Infrastructure
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Write the Goose migration files that introduce all six Phase 3 tables and three new enum types defined in `docs/ADR-003-investment-data-model.md`. Migrations are spread across five sequentially-numbered files, grouped by logical dependency. All files must follow the project's `embed.FS` pattern with `-- +goose Up` / `-- +goose Down` sections.

---

## 2. Context & Motivation

The project uses Goose with an embedded file system for automatic schema migrations at startup (see `internal/platform/db/migrations/`). Phase 3 introduces two global/reference tables (`assets`, `tenant_asset_configs`) and four tenant-scoped tables (`positions`, `position_snapshots`, `position_income_events`, `portfolio_snapshots`) plus three new enums (`asset_type`, `income_type`, `receivable_status`). Every table and index must be present from the first migration so that integration tests run against a correct schema.

The last Phase 2 migration is `00009_create_audit_logs.sql`. Phase 3 migrations start at `00010`.

---

## 3. Scope

### In scope

- [ ] `00010_investment_enums.sql` — `asset_type`, `income_type`, `receivable_status` enums.
- [ ] `00011_investment_global.sql` — `assets` table + `tenant_asset_configs` table.
- [ ] `00012_investment_positions.sql` — `positions` table (FKs to `assets`, `accounts`, `tenants`).
- [ ] `00013_investment_income_events.sql` — `position_snapshots` + `position_income_events` tables.
- [ ] `00014_investment_portfolio_snapshots.sql` — `portfolio_snapshots` table.
- [ ] All indexes, check constraints, and unique partial indexes as per ADR §3.
- [ ] `DOWN` section in each file (DROP TABLE / DROP TYPE in reverse-dependency order).
- [ ] Files placed in `internal/platform/db/migrations/` and picked up automatically by `embed.FS`.

### Out of scope

- Seed data for the `assets` catalogue (separate data-import task).
- Altering any existing Phase 1/2 table.
- `sqlc generate` (Task 3.4).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                                           | Purpose                                                    |
| ------ | ------------------------------------------------------------------------------ | ---------------------------------------------------------- |
| CREATE | `internal/platform/db/migrations/00010_investment_enums.sql`                  | Three new enum types                                       |
| CREATE | `internal/platform/db/migrations/00011_investment_global.sql`                 | `assets` + `tenant_asset_configs`                          |
| CREATE | `internal/platform/db/migrations/00012_investment_positions.sql`              | `positions` table                                          |
| CREATE | `internal/platform/db/migrations/00013_investment_income_events.sql`          | `position_snapshots` + `position_income_events`            |
| CREATE | `internal/platform/db/migrations/00014_investment_portfolio_snapshots.sql`    | `portfolio_snapshots` table                                |

### Schema reference

All DDL is derived directly from ADR-003 §3. Key snippets:

```sql
-- 00010_investment_enums.sql
-- +goose Up
CREATE TYPE asset_type AS ENUM (
  'stock', 'bond', 'fund', 'crypto', 'real_estate', 'income_source'
);
CREATE TYPE income_type AS ENUM (
  'none', 'dividend', 'coupon', 'rent', 'interest', 'salary'
);
CREATE TYPE receivable_status AS ENUM ('pending', 'received', 'cancelled');

-- +goose Down
DROP TYPE IF EXISTS receivable_status;
DROP TYPE IF EXISTS income_type;
DROP TYPE IF EXISTS asset_type;
```

```sql
-- 00011_investment_global.sql
-- +goose Up
CREATE TABLE IF NOT EXISTS assets (
    id          VARCHAR(26)  NOT NULL PRIMARY KEY,
    ticker      VARCHAR(20)  NOT NULL,
    isin        VARCHAR(12),
    name        VARCHAR(200) NOT NULL,
    asset_type  asset_type   NOT NULL,
    currency    CHAR(3)      NOT NULL,
    details     TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_assets_ticker UNIQUE (ticker)
);

CREATE TABLE IF NOT EXISTS tenant_asset_configs (
    id          VARCHAR(26)  NOT NULL PRIMARY KEY,
    tenant_id   VARCHAR(26)  NOT NULL REFERENCES tenants(id),
    asset_id    VARCHAR(26)  NOT NULL REFERENCES assets(id),
    name        VARCHAR(200),
    currency    CHAR(3),
    details     TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    CONSTRAINT uq_tenant_asset_config UNIQUE (tenant_id, asset_id)
        WHERE deleted_at IS NULL
);
CREATE INDEX idx_tenant_asset_configs_tenant
    ON tenant_asset_configs(tenant_id, asset_id);
CREATE INDEX idx_tenant_asset_configs_deleted
    ON tenant_asset_configs(tenant_id, deleted_at);

-- +goose Down
DROP TABLE IF EXISTS tenant_asset_configs;
DROP TABLE IF EXISTS assets;
```

```sql
-- 00012_investment_positions.sql (excerpt)
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
    deleted_at            TIMESTAMPTZ,
    CONSTRAINT uq_position UNIQUE (tenant_id, asset_id, account_id)
        WHERE deleted_at IS NULL
);
-- indexes omitted for brevity — see ADR §3.3

-- +goose Down
DROP TABLE IF EXISTS positions;
```

> For the remaining two files (`00013`, `00014`) derive the DDL verbatim from ADR §3.4 and §3.5–3.6 respectively. Use
> the same naming conventions, FK references, check constraints, and partial-unique indexes shown in the ADR.

```sql
-- 00013 creates: position_snapshots, position_income_events
-- 00014 creates: portfolio_snapshots (with JSONB details column)
```

    id               VARCHAR(26)    NOT NULL PRIMARY KEY,
    tenant_id        VARCHAR(26)    NOT NULL REFERENCES tenants(id),
    asset_id         VARCHAR(26)    NOT NULL REFERENCES assets(id),
    account_id       VARCHAR(26)    NOT NULL REFERENCES accounts(id),
    quantity         NUMERIC(18,8)  NOT NULL DEFAULT 0,
    avg_cost_cents   BIGINT         NOT NULL DEFAULT 0,
    last_price_cents BIGINT         NOT NULL DEFAULT 0,
    currency         CHAR(3)        NOT NULL,
    purchased_at     TIMESTAMPTZ    NOT NULL,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX idx_positions_tenant_id ON positions (tenant_id);
CREATE INDEX idx_positions_asset_id  ON positions (asset_id);

-- +goose Down
DROP INDEX IF EXISTS idx_positions_asset_id;
DROP INDEX IF EXISTS idx_positions_tenant_id;
DROP TABLE IF EXISTS positions;

```

```sql
-- +goose Up

CREATE TABLE IF NOT EXISTS portfolio_snapshots (
    id                 VARCHAR(26)  NOT NULL PRIMARY KEY,
    tenant_id          VARCHAR(26)  NOT NULL REFERENCES tenants(id),
    snapshot_date      DATE         NOT NULL,
    total_value_cents  BIGINT       NOT NULL DEFAULT 0,
    currency           CHAR(3)      NOT NULL,
    details            JSONB,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_snapshot_tenant_date UNIQUE (tenant_id, snapshot_date)
);

CREATE INDEX idx_portfolio_snapshots_tenant_id ON portfolio_snapshots (tenant_id, snapshot_date);

-- +goose Down
DROP INDEX IF EXISTS idx_portfolio_snapshots_tenant_id;
DROP TABLE IF EXISTS portfolio_snapshots;
```

---

## 5. Acceptance Criteria

- [ ] Five migration files exist under `internal/platform/db/migrations/` with numbers `00010`–`00014`.
- [ ] All three enum types (`asset_type`, `income_type`, `receivable_status`) are created in `00010`.
- [ ] `assets` table: no `tenant_id`, correct UNIQUE constraint on `ticker`, `details TEXT` column present.
- [ ] `tenant_asset_configs` table: `tenant_id` FK, partial UNIQUE `(tenant_id, asset_id) WHERE deleted_at IS NULL`, both indexes present.
- [ ] `positions` table: all capital + income-schedule columns present, partial UNIQUE, correct check constraints.
- [ ] `position_snapshots` table: UNIQUE `(position_id, snapshot_date)`, no `deleted_at`.
- [ ] `position_income_events` table: `status receivable_status`, `amount_cents > 0` check, no `deleted_at`.
- [ ] `portfolio_snapshots` table: `total_income_cents BIGINT`, JSONB `details`, UNIQUE `(tenant_id, snapshot_date)`.
- [ ] `goose up` + `goose down` execute without errors against a fresh Postgres container.
- [ ] `make task-check` passes.
- [ ] `docs/ROADMAP.md` row 3.2 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-13 | —      | Task created; updated for ADR v3 (6 tables, 3 enums) |
