# Task 2.7 — Service: `InvoiceCloser` — Scheduled / On-Demand Invoice Closing

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › Service Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-12
> **Assignee:** —
> **Estimated Effort:** L

---

## 1. Summary

Implement `internal/service/invoice_closer.go`, the service that materialises one instalment `transaction` for each open `MasterPurchase` whose current billing cycle is due. It is triggered either on-demand via the API (Task 2.8) or by a future scheduler. Each execution runs entirely within a database transaction to guarantee atomicity: if the transaction `INSERT` fails, `IncrementPaidInstallments` is not committed.

---

## 2. Context & Motivation

The "Ghost Transaction" model requires a dedicated service that "closes" the invoice: reads all pending master purchases for a tenant/account, projects the current instalment amount (Task 2.5 `ProjectInstallments`), inserts the concrete `transactions` row, and increments `paid_installments` on the master purchase. If `paid_installments` reaches `installment_count`, the status flips to `closed` automatically (handled by the atomic `IncrementPaidInstallments` repo call).

The `SYSTEM` actor audit trail (Task 2.10) is also wired here: transactions created by the closer must carry `actor_id = "SYSTEM"` in the audit log.

---

## 3. Scope

### In scope

- [ ] `InvoiceCloser` struct with constructor and `CloseInvoice` method.
- [ ] `CloseInvoice(ctx context.Context, tenantID, accountID string, closingDate time.Time)` — closes all due master purchases for the given account up to `closingDate`.
- [ ] Each materialised instalment is inserted as a `domain.Transaction` with `MasterPurchaseID` set.
- [ ] Each `IncrementPaidInstallments` call wrapped in the same pgx transaction as the `INSERT`.
- [ ] `SYSTEM` actor recorded in audit log for each auto-generated transaction (see Task 2.10).
- [ ] Unit tests with mocked repositories.

### Out of scope

- HTTP handler / trigger endpoint (Task 2.8).
- Audit log repository (already implemented in Phase 1); this task only calls it.
- Scheduler (cron/ticker) — deferred to Phase 5 observability hardening.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                        | Purpose                              |
| ------ | ------------------------------------------- | ------------------------------------ |
| CREATE | `internal/service/invoice_closer.go`        | Invoice closing business logic       |
| CREATE | `internal/service/invoice_closer_test.go`   | Unit tests with mocked dependencies  |

### Key types and methods

```go
// InvoiceCloser materialises instalment transactions at invoice-close time.
type InvoiceCloser struct {
    mpRepo          domain.MasterPurchaseRepository
    txRepo          domain.TransactionRepository
    auditRepo       domain.AuditRepository
    mpSvc           domain.MasterPurchaseService // for ProjectInstallments
    db              *pgxpool.Pool                // for pgx.Tx boundary
}

// NewInvoiceCloser creates a new InvoiceCloser.
func NewInvoiceCloser(
    mpRepo    domain.MasterPurchaseRepository,
    txRepo    domain.TransactionRepository,
    auditRepo domain.AuditRepository,
    mpSvc     domain.MasterPurchaseService,
    db        *pgxpool.Pool,
) *InvoiceCloser

// CloseInvoiceResult reports what the closer did for observability.
type CloseInvoiceResult struct {
    ProcessedCount int
    Errors         []error
}

// CloseInvoice finds all open master purchases for the account due on or before
// closingDate, materialises the current instalment as a transaction, and advances
// paid_installments. Runs each master purchase in its own DB transaction.
func (c *InvoiceCloser) CloseInvoice(
    ctx         context.Context,
    tenantID    string,
    accountID   string,
    closingDate time.Time,
) (CloseInvoiceResult, error)
```

### Closing algorithm (per master purchase)

```
For each pending MasterPurchase:
  1. Call mpSvc.ProjectInstallments(mp) → get full schedule
  2. Take instalments[mp.PaidInstallments] → current instalment
  3. BEGIN DB transaction
  4.   INSERT transactions (amount=instalment.AmountCents, master_purchase_id=mp.ID, type=expense)
  5.   INSERT audit_logs (actor_id="SYSTEM", action="CREATE", entity="transaction", ...)
  6.   CALL IncrementPaidInstallments(tenantID, mp.ID)
  7. COMMIT
  8. If any step fails → ROLLBACK, record error, continue to next purchase
```

### Error cases to handle

| Scenario                                        | Handling                                            |
| ----------------------------------------------- | --------------------------------------------------- |
| `ListPendingClose` returns empty list           | Return `CloseInvoiceResult{ProcessedCount: 0}`, nil |
| DB transaction fails on one master purchase     | Rollback that purchase; continue; aggregate error   |
| `MasterPurchase.PaidInstallments >= InstallmentCount` | Skip (already closed); log warning              |
| Account not found or wrong type                 | Return `domain.ErrNotFound`                         |

---

## 5. Acceptance Criteria

- [ ] All exported functions have Go doc comments.
- [ ] Each instalment materialisation is wrapped in an individual pgx transaction.
- [ ] A DB failure on one master purchase does not abort the others (best-effort, errors aggregated).
- [ ] Materialised transactions have `MasterPurchaseID` set and `Type=expense`.
- [ ] Audit log entry is created for each transaction with `actor_id = "SYSTEM"`.
- [ ] `IncrementPaidInstallments` is called within the same DB transaction as the `INSERT`.
- [ ] Unit tests cover happy path, partial failure, empty pending list.
- [ ] Test coverage for new code ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 2.7 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                     | Type     | Status       |
| ---------------------------------------------- | -------- | ------------ |
| Task 2.1 — `domain/master_purchase.go`         | Upstream | 🔵 backlog   |
| Task 2.4 — `MasterPurchaseRepository`          | Upstream | 🔵 backlog   |
| Task 2.5 — `MasterPurchaseService.ProjectInstallments` | Upstream | 🔵 backlog |
| `domain.TransactionRepository` (Phase 1)       | Upstream | ✅ done      |
| `domain.AuditRepository` (Phase 1)             | Upstream | ✅ done      |
| Task 2.8 — HTTP trigger endpoint (consumer)    | Downstream | 🔵 backlog |
| Task 2.10 — SYSTEM audit actor (same task)     | Inline   | 🔵 backlog   |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/service/invoice_closer_test.go`
- **Cases:**
  - No pending master purchases → `ProcessedCount=0`, nil error.
  - Single pending purchase → one transaction inserted, `IncrementPaidInstallments` called, audit logged.
  - Multiple pending purchases, all succeed → `ProcessedCount=N`, nil error.
  - DB transaction fails on second purchase → first committed, second rolled back, error returned in `Errors` slice.
  - `PaidInstallments >= InstallmentCount` → purchase skipped.
  - Last instalment: `PaidInstallments + 1 == InstallmentCount` → status flips to `closed`.

### Integration tests (`//go:build integration`)

Covered by Task 2.12 — full invoice closing flow integration test.

---

## 8. Open Questions

| # | Question                                                                                     | Owner | Resolution |
| - | -------------------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `CloseInvoice` also recalculate the credit card account balance after each insert?    | —     | Yes — call `AccountRepository.UpdateBalance` within the same DB transaction. |
| 2 | Should the SYSTEM actor have a fixed ULID or the literal string `"SYSTEM"`?                  | —     | Use the string constant `"SYSTEM"` — see Task 2.10. |
| 3 | Should partial failures be logged as structured log entries (slog) in addition to returning errors? | — | Yes — log each failure at `WARN` level with `master_purchase_id` and error. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
