# Task 1.3.9 — Repository Integration Tests (Testcontainers)

> **Roadmap Ref:** Phase 1 — MVP › 1.3 Repository Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** L

---

## 1. Summary

Write integration tests for all Phase 1 repository implementations (Tasks 1.3.1 – 1.3.8) using `testcontainers-go`. Each test suite (file) spins up a fresh Postgres container using the `testutil/containers` package, applies Goose migrations, and provides a clean environment for testing database interactions.

---

## 2. Context & Motivation

Repository implementations cannot be meaningfully unit-tested with mocks — they exist precisely to translate Go calls into SQL. Integration tests against a real Postgres instance (via Testcontainers) are the only way to validate query correctness, constraint enforcement, and error mapping. Each repository will have its own dedicated integration test file to ensure isolation and discoverability.

---

## 3. Scope

### In scope

- [x] Dedicated integration test file for `TenantRepository` (Task 1.3.1) using `containers.NewPostgresDB(t)`.
- [x] Integration tests for `UserRepository` (Task 1.3.2).
- [x] Integration tests for `AuthRepository` / OTP lifecycle (Task 1.3.3).
- [x] Integration tests for `AccountRepository` (Task 1.3.4).
- [x] Integration tests for `CategoryRepository` (Task 1.3.5).
- [x] Integration tests for `TransactionRepository` (Task 1.3.6).
- [x] Integration tests for `AuditRepository` (Task 1.3.7).
- [x] Integration tests for admin repositories (Task 1.3.8).
- [x] All test files use build tag `//go:build integration`.

### Out of scope

- Service-layer tests (Tasks 1.4.x).
- HTTP-level tests (Tasks 1.5.x).
- Load/performance testing (Phase 5).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                               | Purpose                                  |
| ------ | ------------------------------------------------------------------ | ---------------------------------------- |
| CREATE | `internal/platform/repository/tenant_repo_integration_test.go`    | Tenant repository integration tests      |
| CREATE | `internal/platform/repository/user_repo_integration_test.go`      | User repository integration tests        |
| CREATE | `internal/platform/repository/auth_repo_integration_test.go`      | Auth/OTP repository integration tests    |
| CREATE | `internal/platform/repository/account_repo_integration_test.go`   | Account repository integration tests     |
| CREATE | `internal/platform/repository/category_repo_integration_test.go`  | Category repository integration tests    |
| CREATE | `internal/platform/repository/transaction_repo_integration_test.go` | Transaction repository integration tests |
| CREATE | `internal/platform/repository/audit_repo_integration_test.go`     | Audit log repository integration tests   |
| CREATE | `internal/platform/repository/admin_repo_integration_test.go`     | Admin repository integration tests       |

### Key patterns

```go
//go:build integration

package repository_test

func TestExampleRepo_Integration(t *testing.T) {
    t.Parallel()

    ctx := context.Background()

    // Start a fresh container per test suite using the testutil/containers package
    db := containers.NewPostgresDB(t)
    
    // Initialize repository with queries from the container
    repo := repository.NewExampleRepository(db.Queries)

    t.Run("Action", func(t *testing.T) {
        // test logic
    })
}
```

### Test cases per repository

**TenantRepository:**

- Create tenant → GetByID succeeds.
- GetByID with unknown ID → `ErrNotFound`.
- Create duplicate name → `ErrConflict`.
- Delete → soft-deleted; no longer in `List`.

**UserRepository:**

- Create user → GetByID scoped to tenant.
- GetByEmail returns user across tenants.
- Cross-tenant GetByID → `ErrNotFound`.
- UpdateLastLogin updates timestamp.

**AuthRepository:**

- CreateOTPRequest → GetActiveOTPRequest returns it.
- Expired OTP → GetActiveOTPRequest returns `ErrInvalidOTP`.
- MarkOTPUsed → GetActiveOTPRequest returns `ErrInvalidOTP`.
- DeleteExpiredOTPRequests removes expired rows only.

**AccountRepository:**

- Create → GetByID, ListByTenant, ListByUser.
- UpdateBalance → new balance persisted.
- Cross-tenant lookup → `ErrNotFound`.
- SoftDelete → absent from list.

**CategoryRepository:**

- Create root and child → ListChildren returns child only.
- Cross-tenant lookup → `ErrNotFound`.

**TransactionRepository:**

- Create → GetByID returns correct values.
- List with date filter → returns only matching rows.
- List with account filter → returns only that account's rows.
- AmountCents is stored and retrieved as exact int64.

**AuditRepository:**

- Create → ListByEntity returns it.
- ListByTenant with action filter → returns correct subset.

**AdminRepositories:**

- AdminListAllTenants ignores tenant_id → returns rows from multiple tenants.
- AdminHardDelete permanently removes row.

---

## 5. Acceptance Criteria

- [x] `//go:build integration` tag on all test files.
- [x] Testcontainers spins up and tears down a clean Postgres instance via `NewPostgresDB`.
- [x] Migrations are applied via `embed.FS` (handled by utility).
- [x] All repository methods have at least one positive test (happy path).
- [x] Error translation tests: `ErrNotFound`, `ErrConflict`, `ErrInvalidOTP` verified.
- [x] Tests pass with `go test -tags=integration ./internal/platform/repository/...`.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                              | Type     | Status     |
| --------------------------------------- | -------- | ---------- |
| Task 1.3.1 — `tenant_repo.go`           | Upstream | ✅ done |
| Task 1.3.2 — `user_repo.go`             | Upstream | ✅ done |
| Task 1.3.3 — `auth_repo.go`             | Upstream | ✅ done |
| Task 1.3.4 — `account_repo.go`          | Upstream | ✅ done |
| Task 1.3.5 — `category_repo.go`         | Upstream | ✅ done |
| Task 1.3.6 — `transaction_repo.go`      | Upstream | ✅ done |
| Task 1.3.7 — `audit_repo.go`            | Upstream | ✅ done |
| Task 1.3.8 — `admin_repo.go`            | Upstream | ✅ done |
| Task 1.1.6 — Goose migration files      | Upstream | ✅ done    |
| `testcontainers-go` in `vendor/`        | External | ✅ present |

---

## 7. Testing Plan

These are the integration tests themselves — no further testing layer above this for the repository tier.

---

## 8. Open Questions

| # | Question                                                                            | Owner | Resolution |
| - | ----------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should each repo have its own test file or one combined file per repo?              | —     | One file per repository for discoverability. |
| 2 | Should we use a shared TestMain container or fresh per suite?                       | —     | Fresh container per test suite using `testutil/containers.NewPostgresDB(t)`. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
