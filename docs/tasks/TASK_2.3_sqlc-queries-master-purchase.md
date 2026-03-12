# Task 2.3 — sqlc Queries: `master_purchases`

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › Infrastructure
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-12
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Write the raw SQL query file `internal/platform/db/queries/master_purchases.sql` for all `MasterPurchaseRepository` operations and run `sqlc generate` to produce the corresponding Go code. Follows the same pattern as Phase 1 query files.

---

## 2. Context & Motivation

`sqlc` generates type-safe Go code from annotated SQL. Every query must include `WHERE tenant_id = @tenant_id` and `AND deleted_at IS NULL` to enforce multi-tenancy and soft-delete semantics (project rule). The generated `Querier` interface and structs must be usable by the concrete repository (Task 2.4) and mockable in unit tests via the centralized mocks (`internal/testutil/mocks`).

---

## 3. Scope

### In scope

- [ ] SQL file with all required named queries (see §4).
- [ ] `sqlc generate` runs without errors.
- [ ] Generated code committed to `internal/platform/db/sqlc/` (or equivalent generated output path).
- [ ] `MockQuerier` in `internal/testutil/mocks` updated to include new methods.

### Out of scope

- Concrete repository implementation (Task 2.4).
- Business logic (Task 2.5).
- Goose migration (Task 2.2).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                      | Purpose                            |
| ------ | --------------------------------------------------------- | ---------------------------------- |
| CREATE | `internal/platform/db/queries/master_purchases.sql`       | Named sqlc queries                 |
| MODIFY | `internal/testutil/mocks/mock_querier.go`                 | Add mock methods for new queries   |

### SQL queries (sqlc)

```sql
-- name: CreateMasterPurchase :one
INSERT INTO master_purchases (
    id, tenant_id, account_id, category_id, user_id,
    description, total_amount_cents, installment_count,
    closing_day, first_installment_date
) VALUES (
    @id, @tenant_id, @account_id, @category_id, @user_id,
    @description, @total_amount_cents, @installment_count,
    @closing_day, @first_installment_date
)
RETURNING *;

-- name: GetMasterPurchaseByID :one
SELECT * FROM master_purchases
WHERE id = @id
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL;

-- name: ListMasterPurchasesByTenant :many
SELECT * FROM master_purchases
WHERE tenant_id = @tenant_id
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListMasterPurchasesByAccount :many
SELECT * FROM master_purchases
WHERE tenant_id = @tenant_id
  AND account_id = @account_id
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: ListMasterPurchasesPendingClose :many
-- Returns open master purchases where the next instalment due date <= cutoff_date.
SELECT * FROM master_purchases
WHERE tenant_id = @tenant_id
  AND status = 'open'
  AND deleted_at IS NULL
  AND (first_installment_date + (paid_installments * INTERVAL '1 month')) <= @cutoff_date;

-- name: UpdateMasterPurchase :one
UPDATE master_purchases
SET
    category_id  = COALESCE(@category_id, category_id),
    description  = COALESCE(@description, description),
    updated_at   = NOW()
WHERE id = @id
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL
RETURNING *;

-- name: IncrementPaidInstallments :exec
UPDATE master_purchases
SET
    paid_installments = paid_installments + 1,
    status = CASE
        WHEN paid_installments + 1 >= installment_count THEN 'closed'::master_purchase_status
        ELSE status
    END,
    updated_at = NOW()
WHERE id = @id
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL;

-- name: SoftDeleteMasterPurchase :exec
UPDATE master_purchases
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = @id
  AND tenant_id = @tenant_id
  AND deleted_at IS NULL;
```

### Error cases to handle

| Scenario                          | sqlc Behavior                 | Repository handling         |
| --------------------------------- | ----------------------------- | --------------------------- |
| `GetMasterPurchaseByID` no rows   | `pgx.ErrNoRows`               | → `domain.ErrNotFound`      |
| FK violation on `INSERT`          | `pgerrcode.ForeignKeyViolation` | → `domain.ErrNotFound`    |

---

## 5. Acceptance Criteria

- [ ] All queries include `WHERE tenant_id = @tenant_id`.
- [ ] All read queries include `AND deleted_at IS NULL`.
- [ ] `sqlc generate` succeeds with zero errors or warnings.
- [ ] Generated Go types map `total_amount_cents` and `paid_installments` to `int64`/`int32` (no float).
- [ ] `MockQuerier` includes stub methods for all 8 new queries.
- [ ] `golangci-lint run ./...` passes on generated code (generated files may be excluded via `.golangci.yml`).
- [ ] CI `sqlc-diff` check passes (no diff between SQL and generated code).
- [ ] `docs/ROADMAP.md` row 2.3 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                   | Type     | Status       |
| -------------------------------------------- | -------- | ------------ |
| Task 2.1 — `domain/master_purchase.go`       | Upstream | 🔵 backlog   |
| Task 2.2 — Goose migration (table exists)    | Upstream | 🔵 backlog   |
| Task 2.4 — Repository impl (consumer)        | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A — sqlc query files themselves are SQL; correctness is verified via integration tests and the mock contract.

### Integration tests (`//go:build integration`)

Covered by Task 2.4 repository integration tests: each query is exercised against a real Postgres container.

---

## 8. Open Questions

| # | Question                                                                               | Owner | Resolution |
| - | -------------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | `ListMasterPurchasesPendingClose`: should it accept a per-tenant `cutoff_date` or be global? | — | Per-tenant for now; global scheduler can call it per tenant in a loop. |
| 2 | Should `UpdateMasterPurchase` allow updating `closing_day`?                            | —     | No — changing the closing day after creation would invalidate past projections. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
