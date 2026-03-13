# Task 3.4 — SQLC Query Files: `assets` + `tenant_asset_configs`

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Data Access
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Write the raw SQL query files for the `assets` and `tenant_asset_configs` tables and run `sqlc generate` to produce type-safe Go code. Includes the key COALESCE merge query that returns a global asset with any tenant-specific overrides applied (ADR §2.7).

---

## 2. Context & Motivation

The project uses `sqlc` for all database access — no ORM. Before the repositories in Task 3.7 can be implemented, the named queries must exist and `sqlc generate` must have run successfully. The generated code is the contract the repository layer is compiled against.

The `tenant_asset_configs` queries include the COALESCE merge pattern from ADR §2.7: `GetAssetWithTenantConfig` returns a single row where tenant overrides win via `COALESCE(tac.name, a.name)`.

**Reference:** ADR-003 §2.1, §2.7, §3.1, §3.2.

---

## 3. Scope

### In scope

- [x] `internal/platform/db/queries/assets.sql` — CRUD named queries for the global `assets` table.
- [x] `internal/platform/db/queries/tenant_asset_configs.sql` — named queries for tenant overrides including the COALESCE merge query.
- [x] `sqlc generate` runs without errors; generated `.go` files committed.
- [x] `sqlc.yaml` updated if new query files need registration.

### Out of scope

- Queries for `positions`, `position_snapshots`, `position_income_events`, `portfolio_snapshots` (Task 3.11).
- Any analytics / aggregation queries (service-layer computation).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                     | Purpose                                   |
| ------ | -------------------------------------------------------- | ----------------------------------------- |
| CREATE | `internal/platform/db/queries/assets.sql`                | Named queries for global asset catalogue  |
| CREATE | `internal/platform/db/queries/tenant_asset_configs.sql`  | Named queries for tenant overrides        |
| MODIFY | `sqlc.yaml`                                              | Register new query files if needed        |
| MODIFY | `internal/testutil/mocks/querier.go`                   | Manual update of Querier mock interface   |

### Named queries

#### `assets.sql`

```sql
-- name: CreateAsset :one
INSERT INTO assets (id, ticker, isin, name, asset_type, currency, details, created_at)
VALUES (@id, @ticker, @isin, @name, @asset_type, @currency, @details, NOW())
RETURNING *;

-- name: GetAssetByID :one
SELECT * FROM assets
WHERE id = @id;

-- name: GetAssetByTicker :one
SELECT * FROM assets
WHERE ticker = @ticker;

-- name: ListAssets :many
SELECT * FROM assets
ORDER BY ticker ASC;

-- name: DeleteAsset :exec
DELETE FROM assets WHERE id = @id;
```

#### `tenant_asset_configs.sql`

```sql
-- name: UpsertTenantAssetConfig :one
INSERT INTO tenant_asset_configs (id, tenant_id, asset_id, name, currency, details, created_at, updated_at)
VALUES (@id, @tenant_id, @asset_id, @name, @currency, @details, NOW(), NOW())
ON CONFLICT (tenant_id, asset_id) WHERE deleted_at IS NULL
DO UPDATE SET
    name       = EXCLUDED.name,
    currency   = EXCLUDED.currency,
    details    = EXCLUDED.details,
    updated_at = NOW()
RETURNING *;

-- name: GetTenantAssetConfigByAssetID :one
SELECT * FROM tenant_asset_configs
WHERE tenant_id = @tenant_id
  AND asset_id  = @asset_id
  AND deleted_at IS NULL;

-- name: ListTenantAssetConfigs :many
SELECT * FROM tenant_asset_configs
WHERE tenant_id = @tenant_id
  AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: SoftDeleteTenantAssetConfig :exec
UPDATE tenant_asset_configs
SET deleted_at = NOW()
WHERE tenant_id = @tenant_id
  AND asset_id  = @asset_id
  AND deleted_at IS NULL;

-- name: GetAssetWithTenantConfig :one
-- Returns the global asset with tenant overrides applied via COALESCE.
-- Use this query in all asset-display contexts (ADR §2.7).
SELECT
    a.id,
    a.ticker,
    a.isin,
    a.asset_type,
    a.created_at,
    COALESCE(tac.name,     a.name)     AS name,
    COALESCE(tac.currency, a.currency) AS currency,
    COALESCE(tac.details,  a.details)  AS details,
    tac.id         AS config_id,
    tac.created_at AS tenant_added_at
FROM assets a
LEFT JOIN tenant_asset_configs tac
       ON tac.asset_id = a.id
      AND tac.tenant_id = @tenant_id
      AND tac.deleted_at IS NULL
WHERE a.id = @asset_id;
```

---

## 5. Acceptance Criteria

- [x] `assets.sql` contains all five named queries (Create, GetByID, GetByTicker, List, Delete).
- [x] `tenant_asset_configs.sql` contains all five named queries including `GetAssetWithTenantConfig`.
- [x] `GetAssetWithTenantConfig` uses COALESCE for `name`, `currency`, `details` per ADR §2.7.
- [x] Every tenant-scoped query includes `WHERE tenant_id = @tenant_id` and `AND deleted_at IS NULL`.
- [x] `sqlc generate` completes without errors; generated files committed.
- [x] `make task-check` passes.
- [x] `docs/ROADMAP.md` row 3.4 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                                         |
| ---------- | ------ | ---------------------------------------------- |
| 2026-03-13 | —      | Task created; rewritten for ADR v3 (assets + tenant_asset_configs only; COALESCE query added) |
| 2026-03-13 | —      | Implemented SQL queries, ran `sqlc generate`, and manually updated mocks. Passed `make task-check`. |
