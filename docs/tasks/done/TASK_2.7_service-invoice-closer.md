# Task 2.7 — Service: `InvoiceCloser` — Scheduled / On-Demand Invoice Closing

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › Service Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-12
> **Assignee:** GitHub Copilot
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

- [x] `InvoiceCloser` struct with constructor and `CloseInvoice` method.
- [x] `CloseInvoice(ctx context.Context, tenantID, accountID string, closingDate time.Time)` — closes all due master purchases for the given account up to `closingDate`.
- [x] Each materialised instalment is inserted as a `domain.Transaction` with `MasterPurchaseID` set.
- [x] Each `IncrementPaidInstallments` call wrapped in the same pgx transaction as the `INSERT`.
- [x] `SYSTEM` actor recorded in audit log for each auto-generated transaction (see Task 2.10).
- [x] Unit tests with mocked repositories.

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

---

## 5. Acceptance Criteria

- [x] All exported functions have Go doc comments.
- [x] Each instalment materialisation is wrapped in an individual pgx transaction.
- [x] A DB failure on one master purchase does not abort the others (best-effort, errors aggregated).
- [x] Materialised transactions have `MasterPurchaseID` set and `Type=expense`.
- [x] Audit log entry is created for each transaction with `actor_id = "SYSTEM"`.
- [x] `IncrementPaidInstallments` is called within the same DB transaction as the `INSERT`.
- [x] Unit tests cover happy path, partial failure, empty pending list.
- [x] Test coverage for new code ≥ 80%.
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./...` passes with zero issues.
- [x] `docs/ROADMAP.md` row 2.7 updated to ✅ `done`.

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
| 2026-03-12 | GitHub Copilot | Task completed; InvoiceCloser implemented and verified |
