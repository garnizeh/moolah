# Task 2.12 — Integration Tests: Invoice Closing Flow

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › Quality Gate
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-12
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Write end-to-end integration tests that exercise the full invoice closing flow against a real PostgreSQL container: create a credit card account, register a master purchase, trigger `CloseInvoice`, and assert that the correct instalment transaction is materialised, the audit log records `SYSTEM` as actor, and the master purchase status transitions correctly.

---

## 2. Context & Motivation

The `InvoiceCloser` service (Task 2.7) involves multiple DB operations in a single pgx transaction. Unit tests with mocks verify branching logic, but only integration tests against a real DB can catch: constraint failures, correct atomic rollback behaviour, pgx transaction isolation, and the exact state of `master_purchases` and `transactions` tables after closing.

This test file also serves as the acceptance gate for Tasks 2.7, 2.9, and 2.10 combined.

---

## 3. Scope

### In scope

- [ ] Integration test file `internal/service/invoice_closer_integration_test.go` (build tag `integration`).
- [ ] Happy-path: single instalment materialised correctly.
- [ ] Happy-path: all instalments closed → `status=closed`.
- [ ] Remainder-cent invariant verified in DB (last instalment = correct amount).
- [ ] Audit log contains `SYSTEM` actor row.
- [ ] Partial failure: simulate DB error on second master purchase → first committed, second rolled back.
- [ ] Cross-tenant isolation: closing one tenant's account does not affect another tenant's master purchases.

### Out of scope

- HTTP-layer integration (covered by smoke test in Task 1.6.4 pattern; can be extended separately).
- Scheduler/ticker testing (Phase 5).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                          | Purpose                                      |
| ------ | ------------------------------------------------------------- | -------------------------------------------- |
| CREATE | `internal/service/invoice_closer_integration_test.go`         | Integration tests for the closing flow       |
| MODIFY | `internal/testutil/seeds/seeds.go`                            | Add `SeedMasterPurchase` factory helper      |

### Test setup

```go
//go:build integration

package service_test

import (
    "testing"
    "github.com/stretchr/testify/require"
    "github.com/garnizeh/moolah/internal/testutil/containers"
    // ...
)

func TestMain(m *testing.M) {
    containers.SetupPostgres(m) // reuse shared helper from Phase 1
}
```

### Key scenarios

#### Scenario 1 — Single instalment materialised

```
Given: credit_card account, master purchase (total=1200, count=3, paid=0)
When:  CloseInvoice(ctx, tenantID, accountID, today)
Then:  transactions table has 1 new row (amount=400, master_purchase_id=mp.ID)
       master_purchases.paid_installments == 1
       master_purchases.status == "open"
       audit_logs has 1 row with actor_id="SYSTEM"
```

#### Scenario 2 — Final instalment closes purchase

```
Given: master purchase (total=1000, count=3, paid=2)
When:  CloseInvoice(ctx, tenantID, accountID, today)
Then:  transactions amount=334 (remainder absorbed)
       master_purchases.paid_installments == 3
       master_purchases.status == "closed"
```

#### Scenario 3 — No pending purchases

```
Given: master purchase with status="closed"
When:  CloseInvoice(ctx, tenantID, accountID, today)
Then:  ProcessedCount == 0, no new transactions, no error
```

#### Scenario 4 — Cross-tenant isolation

```
Given: tenantA and tenantB each with one pending master purchase
When:  CloseInvoice(ctx, tenantA.ID, accountID, today)
Then:  Only tenantA's master purchase has paid_installments incremented
       tenantB's master purchase unchanged
```

#### Scenario 5 — Remainder-cent invariant

```
Given: master purchase (total=1001, count=3)
When:  Close all 3 instalments sequentially
Then:  Sum of all materialised transaction amounts == 1001
       First 2 transactions == 333 each
       Last transaction == 335
```

### Error cases to handle

| Scenario                        | Expected Behaviour                                    |
| ------------------------------- | ----------------------------------------------------- |
| DB unreachable mid-transaction  | `CloseInvoiceResult.Errors` non-empty; no partial commit |
| `cutoff_date` before all due dates | `ListPendingClose` returns empty; `ProcessedCount=0` |

---

## 5. Acceptance Criteria

- [ ] All 5 scenarios above have passing test cases.
- [ ] Remainder-cent invariant test verifies exact cent values in DB.
- [ ] Audit log `actor_id = "SYSTEM"` verified in DB rows.
- [ ] Cross-tenant isolation verified against a real PG container.
- [ ] `TestMain` uses the shared `testutil/containers.SetupPostgres` helper.
- [ ] All subtests call `t.Parallel()`.
- [ ] Test coverage for new integration code ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 2.12 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                       | Type     | Status       |
| ------------------------------------------------ | -------- | ------------ |
| Task 2.7 — `InvoiceCloser` service               | Upstream | 🔵 backlog   |
| Task 2.9 — Remainder-cent implementation         | Upstream | 🔵 backlog   |
| Task 2.10 — `domain.ActorSystem` constant        | Upstream | 🔵 backlog   |
| Task 2.4 — `MasterPurchaseRepository`            | Upstream | 🔵 backlog   |
| `internal/testutil/containers` (Phase 1)         | Upstream | ✅ done      |
| `internal/testutil/seeds` (Phase 1)              | Upstream | ✅ done      |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A — this task is exclusively integration tests.

### Integration tests (`//go:build integration`)

- **File:** `internal/service/invoice_closer_integration_test.go`
- **Container:** PostgreSQL via `testcontainers-go` (all migrations applied including Task 2.2).
- **Seeds:** `testutil/seeds.SeedMasterPurchase`, `testutil/seeds.SeedCreditCardAccount`.
- **Assertions:** direct DB queries via `pgxpool` to verify row counts and exact field values.

---

## 8. Open Questions

| # | Question                                                                               | Owner | Resolution |
| - | -------------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should this test also exercise the HTTP endpoint (`POST .../close-invoice`) via `httptest`? | — | No — keep HTTP-layer tests in handler unit tests (Task 2.8); this file tests the service directly. |
| 2 | Should `testutil/seeds` add a helper to fast-forward `paid_installments` to N?         | —     | Yes — add `seeds.SetMasterPurchasePaidInstallments(ctx, db, mpID, n)` for Scenario 2. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
