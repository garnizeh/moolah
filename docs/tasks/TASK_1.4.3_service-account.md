# Task 1.4.3 — `service/account_service.go` + unit tests

> **Roadmap Ref:** Phase 1 — MVP › 1.4 Service Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement `AccountService`, which wraps `AccountRepository` with business rules: initial balance seeding on create, balance recalculation guard (balance can only be mutated via `TransactionService`), cross-tenant access prevention, and audit trail writes for every mutating operation.

---

## 2. Context & Motivation

Accounts are central to Phase 1 — every transaction belongs to an account and modifies its balance. The service layer must ensure no direct balance mutations reach the repository outside of transaction processing. See `docs/ARCHITECTURE.md` and roadmap item 1.4.3.

---

## 3. Scope

### In scope

- [ ] `internal/service/account_service.go` — concrete `AccountService` struct.
- [ ] `internal/domain/account.go` — add `AccountService` interface definition.
- [ ] `Create(ctx, tenantID string, input CreateAccountInput) (*Account, error)` — validate, persist, write audit log `create`.
- [ ] `GetByID(ctx, tenantID, id string) (*Account, error)` — fetch with tenant guard.
- [ ] `ListByTenant(ctx, tenantID string) ([]Account, error)` — all accounts for household.
- [ ] `ListByUser(ctx, tenantID, userID string) ([]Account, error)` — accounts for a specific member.
- [ ] `Update(ctx, tenantID, id string, input UpdateAccountInput) (*Account, error)` — metadata only (name, currency); audit log `update`.
- [ ] `Delete(ctx, tenantID, id string) error` — soft-delete; audit log `soft_delete`.
- [ ] Full unit tests in `internal/service/account_service_test.go`.

### Out of scope

- `UpdateBalance` — called only by `TransactionService` (Task 1.4.5); not exposed via `AccountService` interface.
- HTTP handler (`handler/account_handler.go`) — Task 1.5.6.
- Balance recalculation from history — Phase 3+.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                        | Purpose                                       |
| ------ | ------------------------------------------- | --------------------------------------------- |
| MODIFY | `internal/domain/account.go`                | Add `AccountService` interface                |
| CREATE | `internal/service/account_service.go`       | Concrete service implementation               |
| CREATE | `internal/service/account_service_test.go`  | Unit tests with mocked deps                   |

### Key interfaces / types

```go
// AccountService defines the business-logic contract for account management.
type AccountService interface {
    Create(ctx context.Context, tenantID string, input CreateAccountInput) (*Account, error)
    GetByID(ctx context.Context, tenantID, id string) (*Account, error)
    ListByTenant(ctx context.Context, tenantID string) ([]Account, error)
    ListByUser(ctx context.Context, tenantID, userID string) ([]Account, error)
    Update(ctx context.Context, tenantID, id string, input UpdateAccountInput) (*Account, error)
    Delete(ctx context.Context, tenantID, id string) error
}
```

### Business rules

1. `Create`: `input.UserID` must belong to `tenantID` (verify via `UserRepository.GetByID`). Persist, then write audit log `create`.
2. `Update`: fetch existing account first; write audit log `update` with `OldValues`/`NewValues`.
3. `Delete`: ensure no outstanding balance constraints (Phase 1: allow any delete); write audit log `soft_delete`.
4. `UpdateBalance` is intentionally not part of `AccountService` — it is called by `TransactionService` after persisting a transaction, ensuring atomicity within the same DB session (or at the service orchestration level in Phase 1).

### Error cases

| Scenario                         | Sentinel Error         | Notes                          |
| -------------------------------- | ---------------------- | ------------------------------ |
| Account not found                | `domain.ErrNotFound`   | `GetByID`, `Update`, `Delete`  |
| UserID not in tenant             | `domain.ErrForbidden`  | `Create`                       |
| Invalid input                    | validation error       | `Create`, `Update`             |

---

## 5. Acceptance Criteria

- [ ] `AccountService` interface defined in `internal/domain/account.go`.
- [ ] `NewAccountService` constructor accepts `AccountRepository`, `UserRepository`, and `AuditRepository`.
- [ ] `Create` validates that the user belongs to the tenant before persisting.
- [ ] `Update` writes audit log with old and new values.
- [ ] `Delete` writes audit log `soft_delete`.
- [ ] `UpdateBalance` is NOT part of `AccountService`; it is an internal method called by `TransactionService`.
- [ ] Unit tests cover all happy paths and all error branches (≥ 80% coverage).
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                       | Type     | Status  |
| -------------------------------- | -------- | ------- |
| Task 1.3.2 — `user_repo.go`      | Upstream | ✅ done |
| Task 1.3.4 — `account_repo.go`   | Upstream | ✅ done |
| Task 1.3.7 — `audit_repo.go`     | Upstream | ✅ done |
| Task 1.1.17 — `testutil/mocks`   | Upstream | ✅ done |

---

## 7. Testing Plan

Unit tests only.

Key scenarios:

- `Create` → user in tenant → account created, audit log written.
- `Create` → user not in tenant → `ErrForbidden`, no account persisted.
- `GetByID` → not found → `ErrNotFound`.
- `Update` → success → audit log with old/new values.
- `Delete` → success → audit log `soft_delete`.
- `ListByTenant` and `ListByUser` → return expected slices.

---

## 8. Open Questions

| # | Question                                                              | Owner | Resolution |
| - | --------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should currency change be allowed after account creation?             | —     | Yes (Phase 1 only); no conversion logic applied. |
| 2 | Should deleting an account with transactions be blocked?              | —     | No for Phase 1; just soft-delete. Phase 2 revisit. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
