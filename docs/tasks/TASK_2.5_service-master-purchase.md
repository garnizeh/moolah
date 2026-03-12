# Task 2.5 — Service: `MasterPurchaseService` + Unit Tests

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › Service Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-12
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement `internal/service/master_purchase_service.go` satisfying `domain.MasterPurchaseService`. Core responsibilities: validate that the target account is of type `credit_card`, orchestrate repository calls, and implement `ProjectInstallments` — a pure function that computes the full instalment schedule at runtime without any DB writes.

---

## 2. Context & Motivation

The service layer is the single place where business rules are enforced. The key rule for Phase 2 is: **instalments are projected at runtime, never stored until invoice-close time**. The `ProjectInstallments` method embodies this constraint — the `InvoiceCloser` (Task 2.7) consumes this projection to know which instalment to materialise and with what amount. Remainder-cent handling (Task 2.9) is implemented here: the last instalment receives any leftover cents from integer division.

---

## 3. Scope

### In scope

- [ ] `masterPurchaseService` struct implementing `domain.MasterPurchaseService`.
- [ ] Constructor `NewMasterPurchaseService(repo domain.MasterPurchaseRepository, accountRepo domain.AccountRepository) domain.MasterPurchaseService`.
- [ ] `Create`: validates account exists, is `credit_card` type, and belongs to tenant; then delegates to repo.
- [ ] `GetByID`, `ListByTenant`, `ListByAccount`, `Update`, `Delete`: thin orchestration wrappers.
- [ ] `ProjectInstallments(mp *domain.MasterPurchase) []domain.ProjectedInstallment`: pure, deterministic, no side effects.
- [ ] Unit tests with mocked `MasterPurchaseRepository` and `AccountRepository`.

### Out of scope

- `InvoiceCloser` service (Task 2.7) — consumes this service.
- HTTP handlers (Task 2.6).
- Remainder-cent guarantee is implemented here but documented separately in Task 2.9.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                             | Purpose                                      |
| ------ | ------------------------------------------------ | -------------------------------------------- |
| CREATE | `internal/service/master_purchase_service.go`    | Business logic implementation                |
| CREATE | `internal/service/master_purchase_service_test.go` | Unit tests with mocked repositories        |

### Key logic: `ProjectInstallments`

```go
// ProjectInstallments computes each instalment's amount and due date.
// The last instalment absorbs any remainder from integer division to ensure
// the sum of all instalments exactly equals TotalAmountCents.
func ProjectInstallments(mp *domain.MasterPurchase) []domain.ProjectedInstallment {
    base      := mp.TotalAmountCents / int64(mp.InstallmentCount)
    remainder := mp.TotalAmountCents % int64(mp.InstallmentCount)

    instalments := make([]domain.ProjectedInstallment, mp.InstallmentCount)
    for i := range instalments {
        amount := base
        if i == int(mp.InstallmentCount)-1 {
            amount += remainder // last instalment absorbs remainder cents
        }
        instalments[i] = domain.ProjectedInstallment{
            InstallmentNumber: int32(i + 1),
            DueDate:           mp.FirstInstallmentDate.AddDate(0, i, 0),
            AmountCents:       amount,
        }
    }
    return instalments
}
```

### Business rules enforced in `Create`

1. Resolve `AccountID` → account must exist and belong to `tenantID` (`domain.ErrNotFound` otherwise).
2. Account `Type` must be `AccountTypeCreditCard` → return `domain.ErrInvalidInput` otherwise.
3. Delegate to `MasterPurchaseRepository.Create`.

### Error cases to handle

| Scenario                             | Sentinel Error           | HTTP Status |
| ------------------------------------ | ------------------------ | ----------- |
| Account not found / wrong tenant     | `domain.ErrNotFound`     | `404`       |
| Account type is not `credit_card`    | `domain.ErrInvalidInput` | `422`       |
| Master purchase not found            | `domain.ErrNotFound`     | `404`       |
| Delete on closed purchase            | `domain.ErrForbidden`    | `403`       |

---

## 5. Acceptance Criteria

- [ ] All exported functions have Go doc comments.
- [ ] `masterPurchaseService` satisfies `domain.MasterPurchaseService` (compiler check).
- [ ] `create` rejects non-`credit_card` accounts with `domain.ErrInvalidInput`.
- [ ] `ProjectInstallments` sum of all `AmountCents` equals `mp.TotalAmountCents` for any input.
- [ ] `ProjectInstallments` last instalment = `base + remainder` when total is not evenly divisible.
- [ ] `ProjectInstallments` is a pure function (no side effects, no DB calls).
- [ ] Unit tests cover all branches including error paths.
- [ ] Test coverage for new code ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 2.5 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                     | Type     | Status       |
| ---------------------------------------------- | -------- | ------------ |
| Task 2.1 — `domain/master_purchase.go`         | Upstream | 🔵 backlog   |
| Task 2.4 — Repository (interface available)    | Upstream | 🔵 backlog   |
| `domain.AccountRepository` interface           | Upstream | ✅ done      |
| `internal/testutil/mocks` — MockQuerier        | Upstream | ✅ done      |
| Task 2.6 — Handler (consumer)                  | Downstream | 🔵 backlog |
| Task 2.7 — InvoiceCloser (consumer)            | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/service/master_purchase_service_test.go`
- **Cases:**
  - `Create` happy path: valid credit_card account → purchase created.
  - `Create` error: account not found → `ErrNotFound`.
  - `Create` error: account type is `checking` → `ErrInvalidInput`.
  - `GetByID` happy path and `ErrNotFound` propagation.
  - `ListByTenant` and `ListByAccount` return whatever repo returns.
  - `Update` propagates `ErrNotFound`.
  - `Delete` on open purchase → success; on closed purchase → `ErrForbidden`.
  - `ProjectInstallments` with even division (e.g., 1200 / 3 → [400, 400, 400]).
  - `ProjectInstallments` with remainder (e.g., 1000 / 3 → [333, 333, 334]).
  - `ProjectInstallments` due dates advance by 1 month per instalment.

### Integration tests (`//go:build integration`)

N/A — service layer is fully covered by unit tests with mocks. Integration coverage lives in the repository layer (Task 2.4) and the invoice closing flow (Task 2.12).

---

## 8. Open Questions

| # | Question                                                                        | Owner | Resolution |
| - | ------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `ProjectInstallments` respect weekends/holidays for due dates?           | —     | No — fixed monthly offset; leave holiday adjustment to future phase. |
| 2 | Should `Delete` be allowed on closed purchases?                                 | —     | No — once all instalments are materialised, the record is historical. Return `ErrForbidden`. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
