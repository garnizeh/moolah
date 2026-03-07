# Task 1.1.18 — `internal/testutil/seeds`: Canonical Test-Data Factories

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** GitHub Copilot
> **Estimated Effort:** S

---

## 1. Summary

Create `internal/testutil/seeds`, a package of typed factory functions that insert minimal valid rows into a live test database. Each function accepts a `*testing.T`, a `context.Context`, and a `sqlc.Querier`, calls `require.NoError` internally on any failure, and returns the corresponding `domain.*` entity. All files carry `//go:build integration`. Seeds eliminate repetitive fixture boilerplate from every integration test body and guarantee a consistent, shared definition of "a valid test tenant/user/account/etc." across the codebase.

---

## 2. Context & Motivation

Repository integration tests (Task 1.3.9) and future service integration tests need to insert prerequisite rows before exercising the code under test. Without seeds, every test file copies its own inline fixture creation logic — leading to:

1. **Duplication** — each test file that needs a pre-existing tenant re-implements the insert call.
2. **Inconsistency** — field values differ between files; a bug introduced by an inconsistent default can mask real failures.
3. **Fragility** — when the schema changes (e.g., a new non-nullable column) every copypasted fixture must be updated manually.

Centralising seeds in one package means a schema change requires a single fix point, and every integration test immediately benefits.

Architecture reference: `docs/ARCHITECTURE.md` — Section 11.5 Seed Helpers.
Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.18.
Required by: Task 1.3.9 (repository integration tests), Tasks 1.4.x (service integration tests if added).
Depends on: Task 1.1.16 (`testutil/containers` — provides the `sqlc.Querier` used as input).

---

## 3. Scope

### In scope

- [x] `internal/testutil/seeds/tenant.go` — `CreateTenant(t, ctx, q) domain.Tenant`
  - Inserts a row with sensible defaults (`plan: free`, generated ULID)
  - Accepts an optional `name` override via a functional option or a `TenantOpts` struct
- [x] `internal/testutil/seeds/user.go` — `CreateUser(t, ctx, q, tenantID string) domain.User`
  - Inserts a user with `role: member` by default
  - Accepts optional overrides (role, email, name)
- [x] `internal/testutil/seeds/account.go` — `CreateAccount(t, ctx, q, tenantID, userID string) domain.Account`
  - Inserts a checking account with `balance_cents: 0`, currency `BRL`
  - Accepts optional overrides (account type, currency)
- [x] `internal/testutil/seeds/category.go` — `CreateCategory(t, ctx, q, tenantID string) domain.Category`
  - Inserts a root category (no parent) of type `expense`
  - Accepts optional overrides (type, parent ID)
- [x] `internal/testutil/seeds/transaction.go` — `CreateTransaction(t, ctx, q, tenantID, accountID, categoryID, userID string) domain.Transaction`
  - Inserts a minimal expense transaction with `amount_cents: 100`
  - Accepts optional overrides (type, amount, occurred_at)
- [x] All files carry `//go:build integration` at the top
- [x] Each function calls `t.Helper()` at entry and `require.NoError(t, err)` on every fallible operation

### Out of scope

- Bulk/batch seed helpers — single-row factories are sufficient for Phase 1
- Seeds for Phase 2 entities (`master_purchases`) — added when those tables exist
- Cleanup helpers (truncation, deletion) — tests that need isolation should use transactions or a fresh container via `testutil/containers`

---

## 4. Technical Design

### Files to create / modify

| Action | Path | Purpose |
| --- | --- | --- |
| CREATE | `internal/testutil/seeds/tenant.go` | `CreateTenant` factory |
| CREATE | `internal/testutil/seeds/user.go` | `CreateUser` factory |
| CREATE | `internal/testutil/seeds/account.go` | `CreateAccount` factory |
| CREATE | `internal/testutil/seeds/category.go` | `CreateCategory` factory |
| CREATE | `internal/testutil/seeds/transaction.go` | `CreateTransaction` factory |

### Key interfaces / types

```go
// internal/testutil/seeds/tenant.go
//go:build integration

package seeds

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"

    "github.com/garnizeh/moolah/internal/domain"
    "github.com/garnizeh/moolah/internal/platform/db/sqlc"
    "github.com/garnizeh/moolah/pkg/ulid"
)

// CreateTenant inserts a test tenant and returns the mapped domain.Tenant.
// Defaults: plan=free, name="Test Household".
func CreateTenant(t *testing.T, ctx context.Context, q sqlc.Querier) domain.Tenant {
    t.Helper()
    row, err := q.CreateTenant(ctx, sqlc.CreateTenantParams{
        ID:   ulid.New(),
        Name: "Test Household",
        Plan: sqlc.TenantPlanFree,
    })
    require.NoError(t, err)
    return domain.Tenant{
        ID:        row.ID,
        Name:      row.Name,
        Plan:      domain.TenantPlan(row.Plan),
        CreatedAt: row.CreatedAt.Time,
        UpdatedAt: row.UpdatedAt.Time,
    }
}

// --- user.go ---

// CreateUser inserts a test user for the given tenantID.
// Defaults: role=member, name="Test User", unique email.
func CreateUser(t *testing.T, ctx context.Context, q sqlc.Querier, tenantID string) domain.User

// --- account.go ---

// CreateAccount inserts a test checking account owned by userID within tenantID.
// Defaults: type=checking, currency=BRL, balance_cents=0.
func CreateAccount(t *testing.T, ctx context.Context, q sqlc.Querier, tenantID, userID string) domain.Account

// --- category.go ---

// CreateCategory inserts a root expense category within tenantID.
// Defaults: type=expense, no parent, name="Test Category".
func CreateCategory(t *testing.T, ctx context.Context, q sqlc.Querier, tenantID string) domain.Category

// --- transaction.go ---

// CreateTransaction inserts a minimal expense transaction.
// Defaults: type=expense, amount_cents=100, occurred_at=time.Now().
func CreateTransaction(
    t *testing.T,
    ctx context.Context,
    q sqlc.Querier,
    tenantID, accountID, categoryID, userID string,
) domain.Transaction
```

### Usage example (integration test)

```go
//go:build integration

package repository_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/garnizeh/moolah/internal/platform/repository"
    "github.com/garnizeh/moolah/internal/testutil/seeds"
)

func TestTransactionRepository_Create(t *testing.T) {
    t.Parallel()
    ctx := context.Background()

    tenant  := seeds.CreateTenant(t, ctx, sharedDB.Queries)
    user    := seeds.CreateUser(t, ctx, sharedDB.Queries, tenant.ID)
    account := seeds.CreateAccount(t, ctx, sharedDB.Queries, tenant.ID, user.ID)
    cat     := seeds.CreateCategory(t, ctx, sharedDB.Queries, tenant.ID)

    repo := repository.NewTransactionRepository(sharedDB.Queries)
    tx, err := repo.Create(ctx, tenant.ID, domain.Transaction{
        AccountID:   account.ID,
        CategoryID:  cat.ID,
        UserID:      user.ID,
        AmountCents: 500,
        Type:        domain.TransactionTypeExpense,
        Description: "Coffee",
    })
    require.NoError(t, err)
    assert.Equal(t, int64(500), tx.AmountCents)
}
```

### SQL queries (sqlc)

N/A — seeds call existing sqlc queries; no new queries are added.

### API endpoints (if applicable)

N/A

### Error cases to handle

| Scenario | Handling |
| --- | --- |
| Any `q.*` call returns an error | `require.NoError(t, err)` — test fails immediately with location info |
| Missing required FK (e.g., account without valid tenantID) | Caught by DB constraint; surfaces as `require.NoError` failure with clear message |

---

## 5. Acceptance Criteria

- [x] `internal/testutil/seeds/tenant.go` exists with `//go:build integration`; `CreateTenant` compiles and inserts a valid row.
- [x] `internal/testutil/seeds/user.go` exists; `CreateUser` compiles and inserts a valid row with correct `tenant_id`.
- [x] `internal/testutil/seeds/account.go` exists; `CreateAccount` inserts a valid row.
- [x] `internal/testutil/seeds/category.go` exists; `CreateCategory` inserts a valid row.
- [x] `internal/testutil/seeds/transaction.go` exists; `CreateTransaction` inserts a valid row referencing the seeded account, category, and user.
- [x] All functions call `t.Helper()` as their first statement.
- [x] All returned structs are fully populated `domain.*` types (no zero-value fields that should come from the DB, e.g., `CreatedAt`).
- [x] `go vet -tags integration ./internal/testutil/seeds/...` passes with zero issues.
- [x] `golangci-lint run -tags integration ./internal/testutil/seeds/...` passes with zero issues.
- [x] All exported functions have Go doc comments.
- [x] `docs/ROADMAP.md` row 1.1.18 updated to ✅ `done`.

---

## 6. Change Log

| Date | Author | Change |
| --- | --- | --- |
| 2026-03-08 | — | Task document created |
| 2026-03-07 | GitHub Copilot | Implemented all seed factory functions. |
