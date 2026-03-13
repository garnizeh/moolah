# Task 3.11 — SQLC Query Files: Position Family (`positions`, `position_snapshots`, `position_income_events`, `portfolio_snapshots`)

> **Roadmap Ref:** Phase 3 — Investment Portfolio Tracking › Data Access
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-13
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Write the raw SQL query files for the four position-family tables and run `sqlc generate` to produce type-safe Go code. Every query must include `WHERE tenant_id = @tenant_id` and `AND deleted_at IS NULL` where applicable. Blocking dependency for the repository implementation in Task 3.12.

---

## 2. Context & Motivation

Task 3.4 covered the asset-side queries. This task covers the tenant-scoped investment tables. The most important new queries are:
- `ListPositionsDueIncome`: polls `next_income_at <= NOW()` for the income scheduler goroutine (Task 3.13).
- `UpdatePositionNextIncome`: updates `next_income_at` after the scheduler creates an income event.
- `UpdateIncomeEventStatus`: transitions `pending → received | cancelled` for the receivable lifecycle.

**Reference:** ADR-003 §3.3–3.6; income scheduler §9; receivable lifecycle §10.

---

## 3. Scope

### In scope

- [ ] `internal/platform/db/queries/positions.sql` — CRUD + income-schedule queries.
- [ ] `internal/platform/db/queries/position_snapshots.sql` — create + list queries (append-only).
- [ ] `internal/platform/db/queries/position_income_events.sql` — create + status-update + list queries.
- [ ] `internal/platform/db/queries/portfolio_snapshots.sql` — create + list queries (append-only).
- [ ] `sqlc generate` runs without errors; generated `.go` files committed.

### Out of scope

- Asset / tenant_asset_config queries (Task 3.4).
- Analytical aggregation queries (service-layer computation in Go).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                          | Purpose                              |
| ------ | ------------------------------------------------------------- | ------------------------------------ |
| CREATE | `internal/platform/db/queries/positions.sql`                  | Position CRUD + income schedule      |
| CREATE | `internal/platform/db/queries/position_snapshots.sql`         | Append-only snapshot queries         |
| CREATE | `internal/platform/db/queries/position_income_events.sql`     | Receivables CRUD + status transition |
| CREATE | `internal/platform/db/queries/portfolio_snapshots.sql`        | Append-only portfolio aggregate      |

### Named queries (critical subset)

#### `positions.sql`

```sql
-- name: CreatePosition :one
-- name: GetPositionByID :one
-- name: ListPositionsByTenant :many
-- name: ListPositionsByAccount :many
-- name: ListPositionsDueIncome :many
-- SELECT * FROM positions
-- WHERE income_type != 'none' AND next_income_at <= @before AND deleted_at IS NULL

-- name: UpdatePosition :one
-- name: UpdatePositionNextIncome :exec
-- UPDATE positions SET next_income_at = @next, updated_at = NOW()
-- WHERE id = @id AND tenant_id = @tenant_id

-- name: SoftDeletePosition :exec
```

#### `position_income_events.sql`

```sql
-- name: CreatePositionIncomeEvent :one
-- name: GetPositionIncomeEventByID :one
-- name: ListPositionIncomeEventsByTenant :many
-- name: ListPendingIncomeEvents :many
-- SELECT * FROM position_income_events
-- WHERE tenant_id = @tenant_id AND status = 'pending' ORDER BY due_at ASC

-- name: UpdateIncomeEventStatus :one
-- UPDATE position_income_events
-- SET status = @status, received_at = @received_at
-- WHERE id = @id AND tenant_id = @tenant_id
-- RETURNING *
```

#### `portfolio_snapshots.sql`

```sql
-- name: CreatePortfolioSnapshot :one
-- name: GetPortfolioSnapshotByDate :one
-- SELECT * FROM portfolio_snapshots
-- WHERE tenant_id = @tenant_id AND snapshot_date = @date

-- name: ListPortfolioSnapshots :many
```

---

## 5. Acceptance Criteria

- [ ] All four query files created with the named queries listed above.
- [ ] `ListPositionsDueIncome` does NOT include `tenant_id` filter (scheduler processes all tenants globally).
- [ ] `UpdateIncomeEventStatus` only allows transitioning records owned by `tenant_id`.
- [ ] All active-data queries include `AND deleted_at IS NULL` where the table supports soft-delete.
- [ ] `sqlc generate` completes without errors; generated files committed.
- [ ] `make task-check` passes.
- [ ] `docs/ROADMAP.md` row 3.11 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                |
| ---------- | ------ | --------------------- |
| 2026-03-13 | —      | Task created (new)    |
