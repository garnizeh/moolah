# Task 1.1.16 — `internal/testutil/containers`: Centralized Testcontainer Helpers

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-08
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Create `internal/testutil/containers`, a centralized package of thin wrappers around `testcontainers-go` for PostgreSQL, Redis, and Mailhog. Each helper function spins up an ephemeral container, registers `t.Cleanup` for automatic teardown, and returns a typed handle. All files carry the `//go:build integration` tag so they are never compiled in the regular unit-test run. This package replaces the repeated, per-file container setup currently scattered across three integration test files.

---

## 2. Context & Motivation

Three integration test files currently each spin up their own ephemeral container with inline boilerplate:

| File | Container |
| --- | --- |
| `internal/platform/idempotency/redis_store_integration_test.go` | Redis — per test function |
| `internal/platform/mailer/smtp_mailer_integration_test.go` | Mailhog — per test function |
| `internal/platform/repository/*_test.go` (upcoming 1.3.9) | PostgreSQL — not yet written |

Problems with the current approach:

1. **Duplication** — `testcontainers-go` import + container setup is copy-pasted into every file.
2. **No sharing** — each test function starts a fresh container; a `TestMain`-scoped shared container is faster and closer to production behavior.
3. **No central schema application** — the postgres helper needs to apply migrations exactly once per test run, consistently.

Architecture reference: `docs/ARCHITECTURE.md` — Section 11.3 Container Helpers.
Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.16.
Required by: Task 1.3.9 (repository integration tests), future service integration tests.

---

## 3. Scope

### In scope

- [ ] `internal/testutil/containers/postgres.go` — `NewPostgresDB(t *testing.T) *TestPostgresDB`
  - Starts `postgres:17-alpine` via `testcontainers-go/modules/postgres`
  - Applies `docs/schema.sql` as init script
  - Returns `*TestPostgresDB{Pool *pgxpool.Pool, Queries *sqlc.Queries}`
  - Registers `t.Cleanup` for pool close + container termination
- [ ] `internal/testutil/containers/redis.go` — `NewRedisClient(t *testing.T) *redis.Client`
  - Starts `redis:7-alpine`
  - Returns a connected `*redis.Client`
  - Registers `t.Cleanup`
- [ ] `internal/testutil/containers/mailhog.go` — `NewMailhogServer(t *testing.T) *TestMailhog`
  - Starts `mailhog/mailhog:latest` (generic container)
  - Returns `*TestMailhog{SMTPAddr string, APIAddr string}` (both host:port)
  - Registers `t.Cleanup`
- [ ] All files carry `//go:build integration` at the top
- [ ] Unit tests for the helpers themselves are **not** required (they are integration infrastructure)
- [ ] Document `TestMain`-shared-container usage pattern in each file's package doc comment

### Out of scope

- Migrating existing integration tests to use this package (follow-on work in 1.3.9)
- Container health-check tuning
- Parallel container start (future optimization)

---

## 4. Technical Design

### Files to create / modify

| Action | Path | Purpose |
| --- | --- | --- |
| CREATE | `internal/testutil/containers/postgres.go` | PostgreSQL testcontainer helper |
| CREATE | `internal/testutil/containers/redis.go` | Redis testcontainer helper |
| CREATE | `internal/testutil/containers/mailhog.go` | Mailhog testcontainer helper |

### Key interfaces / types

```go
// internal/testutil/containers/postgres.go
//go:build integration

package containers

import (
    "context"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/stretchr/testify/require"
    tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

    "github.com/garnizeh/moolah/internal/platform/db/sqlc"
)

// TestPostgresDB holds an active pgxpool and the sqlc Queries layer bound to it.
type TestPostgresDB struct {
    Pool    *pgxpool.Pool
    Queries *sqlc.Queries
}

// NewPostgresDB starts an ephemeral PostgreSQL container, applies docs/schema.sql,
// and returns a connected TestPostgresDB. Container and pool are cleaned up when t
// completes.
func NewPostgresDB(t *testing.T) *TestPostgresDB

// --- redis.go ---

// TestRedis wraps a connected redis.Client for test use.
// NewRedisClient starts an ephemeral Redis container and returns a connected
// *redis.Client. Container is cleaned up when t completes.
func NewRedisClient(t *testing.T) *redis.Client

// --- mailhog.go ---

// TestMailhog holds the exposed SMTP and HTTP API addresses of the Mailhog container.
type TestMailhog struct {
    SMTPAddr string // e.g. "localhost:1025"
    APIAddr  string // e.g. "http://localhost:8025"
}

// NewMailhogServer starts an ephemeral Mailhog container and returns the
// TestMailhog handle. Container is cleaned up when t completes.
func NewMailhogServer(t *testing.T) *TestMailhog
```

### Shared-container pattern for test packages

```go
// Example: internal/platform/repository/main_test.go
//go:build integration

package repository_test

import (
    "os"
    "testing"

    "github.com/garnizeh/moolah/internal/testutil/containers"
)

var sharedDB *containers.TestPostgresDB

func TestMain(m *testing.M) {
    t := &testing.T{}
    sharedDB = containers.NewPostgresDB(t)
    os.Exit(m.Run())
}
```

### SQL queries (sqlc)

N/A — this task does not add any sqlc queries.

### API endpoints (if applicable)

N/A

### Error cases to handle

| Scenario | Handling |
| --- | --- |
| Container fails to start | `require.NoError(t, err)` — test is immediately failed with a clear message |
| `ConnectionString` fails | `require.NoError(t, err)` |
| `pgxpool.New` fails | `require.NoError(t, err)` |

---

## 5. Acceptance Criteria

- [ ] `internal/testutil/containers/postgres.go` exists with `//go:build integration` and exports `NewPostgresDB` returning `*TestPostgresDB`.
- [ ] `internal/testutil/containers/redis.go` exists with `//go:build integration` and exports `NewRedisClient` returning `*redis.Client`.
- [ ] `internal/testutil/containers/mailhog.go` exists with `//go:build integration` and exports `NewMailhogServer` returning `*TestMailhog`.
- [ ] `TestPostgresDB.Queries` is a fully initialized `*sqlc.Queries` usable immediately after `NewPostgresDB` returns.
- [ ] `t.Cleanup` is always registered — no manual `Terminate` calls needed by callers.
- [ ] `go build ./...` (without `-tags integration`) compiles without importing this package.
- [ ] `go vet -tags integration ./internal/testutil/containers/...` passes with zero issues.
- [ ] `golangci-lint run -tags integration ./internal/testutil/containers/...` passes with zero issues.
- [ ] All exported types and functions have Go doc comments.
- [ ] `docs/ROADMAP.md` row 1.1.16 updated to ✅ `done`.

---

## 6. Change Log

| Date | Author | Change |
| --- | --- | --- |
| 2026-03-08 | — | Task document created |
