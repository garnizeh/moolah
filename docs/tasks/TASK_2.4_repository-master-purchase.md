# Task 2.4 — Repository: `MasterPurchaseRepository` Implementation + Integration Tests

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › Repository Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-12
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the concrete `masterPurchaseRepo` struct in `internal/platform/repository/master_purchase_repo.go` using the sqlc-generated `Querier`. The implementation must satisfy `domain.MasterPurchaseRepository` and include full integration test coverage using `testcontainers-go`.

---

## 2. Context & Motivation

Following the repository pattern established in Phase 1, this task bridges the sqlc-generated code (Task 2.3) and the domain interface (Task 2.1). All pgx errors must be translated to domain sentinel errors. The `IncrementPaidInstallments` method is critical — it is called inside a transaction by the `InvoiceCloser` service (Task 2.7) so it must be idempotent and atomic.

---

## 3. Scope

### In scope

- [ ] `masterPurchaseRepo` struct implementing `domain.MasterPurchaseRepository`.
- [ ] Constructor `NewMasterPurchaseRepository(q sqlc.Querier) domain.MasterPurchaseRepository`.
- [ ] Mapping functions between `sqlc.MasterPurchase` and `domain.MasterPurchase`.
- [ ] Error translation: `pgx.ErrNoRows` → `domain.ErrNotFound`; FK violation → `domain.ErrNotFound`.
- [ ] Integration tests with `testcontainers-go` covering all methods.

### Out of scope

- Business logic / instalment projection (Task 2.5).
- `InvoiceCloser` service (Task 2.7).
- HTTP handlers (Task 2.6).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                               | Purpose                                        |
| ------ | ------------------------------------------------------------------ | ---------------------------------------------- |
| CREATE | `internal/platform/repository/master_purchase_repo.go`            | Concrete MasterPurchaseRepository implementation |
| CREATE | `internal/platform/repository/master_purchase_repo_test.go`       | Integration tests (build tag: integration)     |

### Key interfaces / types

```go
type masterPurchaseRepo struct {
    q sqlc.Querier
}

// NewMasterPurchaseRepository creates a new MasterPurchaseRepository backed by sqlc.
func NewMasterPurchaseRepository(q sqlc.Querier) domain.MasterPurchaseRepository {
    return &masterPurchaseRepo{q: q}
}
```

### Mapping pattern

```go
func toDomainMasterPurchase(row sqlc.MasterPurchase) *domain.MasterPurchase {
    mp := &domain.MasterPurchase{
        ID:                   row.ID,
        TenantID:             row.TenantID,
        AccountID:            row.AccountID,
        CategoryID:           row.CategoryID,
        UserID:               row.UserID,
        Description:          row.Description,
        Status:               domain.MasterPurchaseStatus(row.Status),
        TotalAmountCents:     row.TotalAmountCents,
        InstallmentCount:     row.InstallmentCount,
        PaidInstallments:     row.PaidInstallments,
        ClosingDay:           row.ClosingDay,
        FirstInstallmentDate: row.FirstInstallmentDate.Time,
        CreatedAt:            row.CreatedAt.Time,
        UpdatedAt:            row.UpdatedAt.Time,
    }
    if row.DeletedAt.Valid {
        t := row.DeletedAt.Time
        mp.DeletedAt = &t
    }
    return mp
}
```

### Error cases to handle

| Scenario                             | pgx Error                       | Domain Error              |
| ------------------------------------ | ------------------------------- | ------------------------- |
| Row not found (Get, Update, Delete)  | `pgx.ErrNoRows`                 | `domain.ErrNotFound`      |
| FK violation (account/category/user) | `pgerrcode.ForeignKeyViolation` | `domain.ErrNotFound`      |
| CHECK violation (cents, installments)| `pgerrcode.CheckViolation`      | `domain.ErrInvalidInput`  |

---

## 5. Acceptance Criteria

- [ ] All exported types and functions have Go doc comments.
- [ ] Struct implements `domain.MasterPurchaseRepository` (verified by compiler assertion).
- [ ] Every method passes `tenant_id` — no cross-tenant data leakage.
- [ ] `AmountCents` and related fields map to `int64`/`int32` with no float conversion.
- [ ] All pgx errors are translated to domain sentinel errors.
- [ ] Integration tests cover Create, GetByID, ListByTenant, ListByAccount, ListPendingClose, Update, IncrementPaidInstallments, Delete.
- [ ] Integration test verifies `IncrementPaidInstallments` auto-closes purchase when `paid == total`.
- [ ] Integration test verifies soft-deleted records are excluded from all list queries.
- [ ] Integration test verifies cross-tenant queries return zero rows.
- [ ] Test coverage for new code ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 2.4 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                               | Type     | Status       |
| ---------------------------------------- | -------- | ------------ |
| Task 2.1 — `domain/master_purchase.go`   | Upstream | 🔵 backlog   |
| Task 2.2 — Goose migration               | Upstream | 🔵 backlog   |
| Task 2.3 — sqlc queries + generated code | Upstream | 🔵 backlog   |
| `internal/testutil/containers`           | Upstream | ✅ done      |
| `internal/testutil/seeds`                | Upstream | ✅ done      |
| Task 2.5 — Service (consumer)            | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A — repository layer is tested exclusively via integration tests against a real DB.

### Integration tests (`//go:build integration`)

- **File:** `internal/platform/repository/master_purchase_repo_test.go`
- **Setup:** `TestMain` uses `internal/testutil/containers.SetupPostgres`; seeds use `testutil/seeds` helpers.
- **Cases:**
  - `Create` persists a valid master purchase and returns it with generated `id`.
  - `GetByID` returns correct record for matching `tenant_id`; returns `ErrNotFound` for wrong tenant.
  - `ListByTenant` returns all non-deleted records; excludes soft-deleted.
  - `ListByAccount` filters by `account_id` within tenant.
  - `ListPendingClose` returns only `open` records where next instalment is due.
  - `Update` changes description/category; returns `ErrNotFound` on missing ID.
  - `IncrementPaidInstallments` increments counter; sets `status=closed` when `paid==total`.
  - `Delete` soft-deletes; subsequent `GetByID` returns `ErrNotFound`.
  - Cross-tenant: all methods return empty/ErrNotFound for foreign `tenant_id`.

---

## 8. Open Questions

| # | Question                                                                           | Owner | Resolution |
| - | ---------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `IncrementPaidInstallments` be wrapped in a DB transaction alongside the transaction INSERT? | — | Yes — the `InvoiceCloser` will call it within a `pgx.Tx`; the repo method should accept a `Querier` (which can be a tx). |
| 2 | Should the integration test for `ListPendingClose` manipulate `paid_installments` directly via SQL? | — | Yes — seed helpers may need a `SetPaidInstallments` utility. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
