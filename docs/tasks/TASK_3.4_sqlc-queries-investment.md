# Task 3.4 — SQLC Query Files for Investment Entities

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Data Access
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Write all raw SQL query files for the Phase 3 investment tables (`assets`, `positions`, `portfolio_snapshots`) in `internal/platform/db/queries/`, then run `sqlc generate` to produce the corresponding type-safe Go code. The generated code will be consumed by the repository implementations in Task 3.5.

---

## 2. Context & Motivation

The project uses `sqlc` to generate Go code from raw SQL — no ORMs. Every query is explicit, auditable, and type-safe. Phase 1 and Phase 2 followed this pattern. Phase 3 must add named query files for all three investment tables before the repository layer can be implemented in Task 3.5.

Key constraints:

- Every tenant-scoped query **must** include `WHERE tenant_id = @tenant_id`.
- Every active-data query **must** include `AND deleted_at IS NULL` for `positions`.
- The `assets` table is global (no `tenant_id` filter).

---

## 3. Scope

### In scope

- [ ] `internal/platform/db/queries/assets.sql` — CRUD queries for global asset catalogue.
- [ ] `internal/platform/db/queries/positions.sql` — tenant-scoped CRUD + list queries.
- [ ] `internal/platform/db/queries/portfolio_snapshots.sql` — create + list queries.
- [ ] `sqlc generate` executed successfully; generated Go files committed.
- [ ] `sqlc.yaml` updated if new query files need explicit registration.

### Out of scope

- Complex analytical queries (e.g., allocation calculations) — these live in the service layer using Go logic over the basic query results.
- Queries for price history (not in Phase 3 scope).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                      | Purpose                               |
| ------ | --------------------------------------------------------- | ------------------------------------- |
| CREATE | `internal/platform/db/queries/assets.sql`                | Named queries for assets              |
| CREATE | `internal/platform/db/queries/positions.sql`             | Named queries for positions           |
| CREATE | `internal/platform/db/queries/portfolio_snapshots.sql`   | Named queries for portfolio snapshots |
| MODIFY | `sqlc.yaml`                                               | Register new query files if needed    |

### Named Queries

#### `assets.sql`

```sql
-- name: CreateAsset :one
INSERT INTO assets (id, ticker, name, asset_type, currency, created_at)
VALUES (@id, @ticker, @name, @asset_type, @currency, NOW())
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
```

#### `positions.sql`

```sql
-- name: CreatePosition :one
INSERT INTO positions (
    id, tenant_id, asset_id, account_id,
    quantity, avg_cost_cents, last_price_cents,
    currency, purchased_at, created_at, updated_at
)
VALUES (
    @id, @tenant_id, @asset_id, @account_id,
    @quantity, @avg_cost_cents, @last_price_cents,
    @currency, @purchased_at, NOW(), NOW()
)
RETURNING *;

-- name: GetPositionByID :one
SELECT * FROM positions
WHERE tenant_id = @tenant_id
  AND id = @id
  AND deleted_at IS NULL;

-- name: ListPositionsByTenant :many
SELECT * FROM positions
WHERE tenant_id = @tenant_id
  AND deleted_at IS NULL
ORDER BY purchased_at DESC;

-- name: ListPositionsByAccount :many
SELECT * FROM positions
WHERE tenant_id = @tenant_id
  AND account_id = @account_id
  AND deleted_at IS NULL
ORDER BY purchased_at DESC;

-- name: UpdatePosition :one
UPDATE positions
SET
    quantity         = COALESCE(@quantity, quantity),
    avg_cost_cents   = COALESCE(@avg_cost_cents, avg_cost_cents),
    last_price_cents = COALESCE(@last_price_cents, last_price_cents),
    updated_at       = NOW()
WHERE tenant_id = @tenant_id
  AND id = @id
  AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeletePosition :exec
UPDATE positions
SET deleted_at = NOW()
WHERE tenant_id = @tenant_id
  AND id = @id
  AND deleted_at IS NULL;
```

#### `portfolio_snapshots.sql`

```sql
-- name: CreatePortfolioSnapshot :one
INSERT INTO portfolio_snapshots (
    id, tenant_id, snapshot_date, total_value_cents, currency, details, created_at
)
VALUES (
    @id, @tenant_id, @snapshot_date, @total_value_cents, @currency, @details, NOW()
)
RETURNING *;

-- name: GetPortfolioSnapshotByDate :one
SELECT * FROM portfolio_snapshots
WHERE tenant_id = @tenant_id
  AND snapshot_date = @snapshot_date;

-- name: ListPortfolioSnapshotsByTenant :many
SELECT * FROM portfolio_snapshots
WHERE tenant_id = @tenant_id
ORDER BY snapshot_date DESC;
```

---

## 5. Acceptance Criteria

- [ ] All query files are present in `internal/platform/db/queries/`.
- [ ] `sqlc generate` runs with zero errors.
- [ ] Generated Go code is committed (`internal/platform/db/sqlc/`).
- [ ] All tenant-scoped queries include `WHERE tenant_id = @tenant_id`.
- [ ] All `positions` queries include `AND deleted_at IS NULL` for active records.
- [ ] `assets` queries do NOT include `tenant_id` (global catalogue).
- [ ] `golangci-lint run ./...` passes with zero issues on generated code (or generated files are excluded in `.golangci.yml`).
- [ ] `docs/ROADMAP.md` row 3.4 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                        | Type     | Status     |
| ------------------------------------------------- | -------- | ---------- |
| Task 3.2 — Migrations create the target tables    | Upstream | 🔵 backlog |
| Task 3.3 — Domain types define expected field names | Upstream | 🔵 backlog |
| `sqlc.yaml` already configured                    | Upstream | ✅ done    |

---

## 7. Testing Plan

### Unit tests

N/A — SQL queries are validated by integration tests.

### Integration tests

- All repository tests in Task 3.5 exercise these queries against a live Postgres container.
- `sqlc generate` output is verified in CI via `make generate && git diff --exit-code`.

---

## 8. Open Questions

| # | Question                                                                                        | Owner | Resolution |
| - | ----------------------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `UpdatePosition` use `COALESCE` nullable params or a separate `sqlc` nullable type?     | —     | Follow Phase 2 pattern (pointer params via `sqlc.narg`) |
| 2 | Does `portfolio_snapshots` need an `UPDATE` query, or is it always insert-only?                | —     | Insert-only for MVP (snapshots are immutable point-in-time records) |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-13 | —      | Task created from roadmap |
