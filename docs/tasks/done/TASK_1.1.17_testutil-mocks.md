# Task 1.1.17 — `internal/testutil/mocks`: Centralized testify/mock Implementations

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Create `internal/testutil/mocks`, a centralized package that houses all reusable `testify/mock` implementations for the project's domain interfaces. The package consolidates three scattered definitions — `MockQuerier` (currently in `internal/platform/db/sqlc/mock_querier.go`), `MockIdempotencyStore` (currently inline in `internal/platform/middleware/idempotency_test.go`), and a new `MockMailer` — into a single, importable location. This makes every mock available to any test file across the codebase without code duplication.

---

## 2. Context & Motivation

Current state of mocks:

| Mock | Current location | Problem |
| --- | --- | --- |
| `MockQuerier` | `internal/platform/db/sqlc/mock_querier.go` — `package sqlc` | Coupled to the generated package; awkward import path for service tests |
| `MockIdempotencyStore` | `internal/platform/middleware/idempotency_test.go` | Trapped in a `_test.go` file; **not importable** by handler tests or integration suites |
| `MockMailer` | Does not exist | Service tests for `AuthService` require a mockable `domain.Mailer` |

Moving them to `internal/testutil/mocks` (as regular `.go` files, no build tag) makes them importable by any test file in the project. The package imports only `testify/mock` and domain/middleware types — no infrastructure clients — so it is safe to compile in every environment including the linter step.

Architecture reference: `docs/ARCHITECTURE.md` — Section 11.4 Mock Implementations.
Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.17.
Required by: Tasks 1.4.x (service layer unit tests), 1.5.x (handler unit tests).

---

## 3. Scope

### In scope

- [ ] `internal/testutil/mocks/mock_querier.go`
  - Move `MockQuerier` from `internal/platform/db/sqlc/mock_querier.go`
  - Change package from `sqlc` to `mocks`
  - All methods import `sqlc` types; compile-time interface check: `var _ sqlc.Querier = (*MockQuerier)(nil)`
- [ ] `internal/testutil/mocks/mock_idempotency_store.go`
  - Extract `MockIdempotencyStore` from `internal/platform/middleware/idempotency_test.go`
  - Make it a standalone, importable type in `package mocks`
  - Compile-time check: `var _ middleware.IdempotencyStore = (*MockIdempotencyStore)(nil)`
- [ ] `internal/testutil/mocks/mock_mailer.go`
  - New `MockMailer` implementing `domain.Mailer`
  - Compile-time check: `var _ domain.Mailer = (*MockMailer)(nil)`
- [ ] Update `internal/platform/db/sqlc/mock_querier.go` — replace with a compile-time re-export or remove and update all import sites
- [ ] Update `internal/platform/middleware/idempotency_test.go` — remove the inline `MockIdempotencyStore` definition, import from `testutil/mocks` instead
- [ ] Update all existing repository tests that import `sqlc.MockQuerier` to use `mocks.Querier`
- [ ] No build tags on any file in this package

### Out of scope

- `MockTenantRepository`, `MockUserRepository`, `MockTransactionRepository` etc. — those will be added as part of the service layer tasks (1.4.x) when their interfaces are exercised in tests
- `gomock` / `moq` code generation (testify/mock is the established pattern; keep consistency)

---

## 4. Technical Design

### Files to create / modify

| Action | Path | Purpose |
| --- | --- | --- |
| CREATE | `internal/testutil/mocks/mock_querier.go` | MockQuerier — implements `sqlc.Querier` |
| CREATE | `internal/testutil/mocks/mock_idempotency_store.go` | MockIdempotencyStore — implements `middleware.IdempotencyStore` |
| CREATE | `internal/testutil/mocks/mock_mailer.go` | MockMailer — implements `domain.Mailer` |
| MODIFY | `internal/platform/db/sqlc/mock_querier.go` | Replace body with type alias pointing to `mocks.Querier`, or delete and fix imports |
| MODIFY | `internal/platform/middleware/idempotency_test.go` | Remove inline `MockIdempotencyStore` definition; import `mocks` instead |
| MODIFY | `internal/platform/repository/*_test.go` | Update import from `sqlc` package to `mocks` package for `MockQuerier` |

### Key interfaces / types

```go
// internal/testutil/mocks/mock_querier.go
package mocks

import (
    "context"

    "github.com/stretchr/testify/mock"

    "github.com/garnizeh/moolah/internal/platform/db/sqlc"
)

// MockQuerier is a testify/mock implementation of sqlc.Querier.
// It centralises the mock so all repository, service, and handler tests can import it.
type MockQuerier struct {
    mock.Mock
}

// ... all Querier method implementations (same logic as current mock_querier.go) ...

// Compile-time interface check.
var _ sqlc.Querier = (*MockQuerier)(nil)

// --- mock_idempotency_store.go ---
package mocks

import (
    "context"
    "time"

    "github.com/stretchr/testify/mock"

    "github.com/garnizeh/moolah/internal/platform/middleware"
)

// MockIdempotencyStore is a testify/mock implementation of middleware.IdempotencyStore.
type MockIdempotencyStore struct {
    mock.Mock
}

func (m *MockIdempotencyStore) Get(ctx context.Context, key string) (*middleware.CachedResponse, error) {
    args := m.Called(ctx, key)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*middleware.CachedResponse), args.Error(1)
}

func (m *MockIdempotencyStore) SetLocked(ctx context.Context, key string, ttl time.Duration) (bool, error) {
    args := m.Called(ctx, key, ttl)
    return args.Bool(0), args.Error(1)
}

func (m *MockIdempotencyStore) SetResponse(ctx context.Context, key string, resp middleware.CachedResponse, ttl time.Duration) error {
    args := m.Called(ctx, key, resp, ttl)
    return args.Error(0)
}

var _ middleware.IdempotencyStore = (*MockIdempotencyStore)(nil)

// --- mock_mailer.go ---
package mocks

import (
    "context"

    "github.com/stretchr/testify/mock"

    "github.com/garnizeh/moolah/internal/domain"
)

// MockMailer is a testify/mock implementation of domain.Mailer.
type MockMailer struct {
    mock.Mock
}

func (m *MockMailer) SendOTP(ctx context.Context, email, code string) error {
    args := m.Called(ctx, email, code)
    return args.Error(0)
}

var _ domain.Mailer = (*MockMailer)(nil)
```

### SQL queries (sqlc)

N/A

### API endpoints (if applicable)

N/A

### Error cases to handle

N/A — mock files contain no business logic.

---

## 5. Acceptance Criteria

- [ ] `internal/testutil/mocks/mock_querier.go` exists in `package mocks` and `MockQuerier` implements every method of `sqlc.Querier` (compile-time check passes).
- [ ] `internal/testutil/mocks/mock_idempotency_store.go` exists; `MockIdempotencyStore` implements `middleware.IdempotencyStore` (compile-time check passes).
- [ ] `internal/testutil/mocks/mock_mailer.go` exists; `MockMailer` implements `domain.Mailer` (compile-time check passes).
- [ ] No file in `internal/testutil/mocks/` carries a `//go:build` tag.
- [ ] `internal/platform/middleware/idempotency_test.go` no longer defines `MockIdempotencyStore` inline — it imports from `testutil/mocks`.
- [ ] All repository `_test.go` files that previously referenced `sqlc.MockQuerier` now reference `mocks.Querier`.
- [ ] `go test ./...` (unit tests, no integration tag) passes with zero failures.
- [ ] `golangci-lint run ./internal/testutil/mocks/...` passes with zero issues.
- [ ] All exported types and functions have Go doc comments.
- [ ] `docs/ROADMAP.md` row 1.1.17 updated to ✅ `done`.

---

## 6. Change Log

| Date | Author | Change |
| --- | --- | --- |
| 2026-03-08 | — | Task document created |
| 2026-03-07 | Automated agent | Implemented centralized mocks in `internal/testutil/mocks`, updated tests to use them, removed legacy `internal/platform/db/sqlc/mock_querier.go` and regenerated sqlc artifacts where needed. |
