# Task 3.2 — Goose Migrations: `assets`, `positions`, `portfolio_snapshots`

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Infrastructure
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Write the Goose migration files that introduce the three Phase 3 tables — `assets`, `positions`, and `portfolio_snapshots` — plus the `asset_type` enum. Migrations must follow the project's established pattern: sequential numbered files using `embed.FS`, with `UP` and `DOWN` sections, and all schema decisions derived from the ADR produced in Task 3.1.

---

## 2. Context & Motivation

The project uses Goose with an embedded file system for automatic schema migrations at startup. All Phase 1 and Phase 2 tables were introduced this way. Phase 3 must follow the same pattern so that CI integration tests and production deployments apply schema changes atomically and reversibly.

The `assets` table is a global reference catalogue (no `tenant_id`). The `positions` table is tenant-scoped. The `portfolio_snapshots` table is also tenant-scoped. Correct FK constraints, indexes, and the soft-delete column must all be present from day one.

---

## 3. Scope

### In scope

- [ ] New Goose migration file: `assets` table + `asset_type` enum.
- [ ] New Goose migration file: `positions` table with FK to `assets` and `accounts`.
- [ ] New Goose migration file: `portfolio_snapshots` table.
- [ ] Indexes: `positions(tenant_id)`, `positions(asset_id)`, `portfolio_snapshots(tenant_id, snapshot_date)`.
- [ ] `DOWN` migration for each file (reversal only; not required to be idempotent).
- [ ] Files placed in `internal/platform/db/migrations/` and embedded via `embed.FS`.

### Out of scope

- Seed data for the `assets` catalogue (deferred; may come from a future data-import task).
- Altering existing Phase 1/2 tables (no cross-phase migration mixing).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                                  | Purpose                                |
| ------ | --------------------------------------------------------------------- | -------------------------------------- |
| CREATE | `internal/platform/db/migrations/000X_investment_assets.sql`         | `asset_type` enum + `assets` table     |
| CREATE | `internal/platform/db/migrations/000Y_investment_positions.sql`      | `positions` table                      |
| CREATE | `internal/platform/db/migrations/000Z_investment_snapshots.sql`      | `portfolio_snapshots` table            |

> Migration numbers must continue sequentially from the last Phase 2 migration file.

### Schema

```sql
-- +goose Up

CREATE TYPE asset_type AS ENUM ('stock', 'bond', 'fund', 'crypto', 'real_estate');

CREATE TABLE IF NOT EXISTS assets (
    id           VARCHAR(26)  NOT NULL PRIMARY KEY,
    ticker       VARCHAR(20)  NOT NULL,
    name         VARCHAR(200) NOT NULL,
    asset_type   asset_type   NOT NULL,
    currency     CHAR(3)      NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_assets_ticker UNIQUE (ticker)
);

-- +goose Down
DROP TABLE IF EXISTS assets;
DROP TYPE IF EXISTS asset_type;
```

```sql
-- +goose Up

CREATE TABLE IF NOT EXISTS positions (
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

- [ ] All three migration files are present and follow the project naming convention.
- [ ] `goose up` applies cleanly against a fresh Postgres instance (verified by integration test containers).
- [ ] `goose down` reverses the migrations without errors.
- [ ] Every Phase 3 table has `tenant_id` (except `assets`, which is global by design, as per ADR 3.1).
- [ ] All foreign key constraints are correct.
- [ ] Soft-delete column (`deleted_at`) is present on `positions`.
- [ ] Indexes are present for tenant-scoped queries.
- [ ] `docs/ROADMAP.md` row 3.2 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                                  | Type     | Status       |
| ----------------------------------------------------------- | -------- | ------------ |
| Task 3.1 — ADR defines table structure                      | Upstream | 🔵 backlog   |
| Last Phase 2 migration number (to pick correct sequence ID) | Upstream | ✅ done      |

---

## 7. Testing Plan

### Unit tests

N/A — migrations are verified by integration tests.

### Integration tests (`//go:build integration`)

- Existing `containers.NewPostgresDB` helper runs all migrations automatically.
- Any Task 3.5 integration test against the repository layer implicitly validates these migrations.
- Optionally add a dedicated smoke test: `goose status` returns all migrations as applied.

---

## 8. Open Questions

| # | Question                                                                    | Owner | Resolution |
| - | --------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `assets` have a soft-delete column, or is hard-delete acceptable?   | —     | Resolve in Task 3.1 ADR |
| 2 | Is `account_id` in `positions` required (nullable FK) or mandatory?        | —     | Resolve in Task 3.1 ADR |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-13 | —      | Task created from roadmap |
