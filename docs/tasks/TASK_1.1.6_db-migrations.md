# Task 1.1.6 — Database Migrations (Goose)

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** � `in-progress`
> **Last Updated:** 2026-03-07
> **Assignee:** GitHub Copilot
> **Estimated Effort:** M

---

## 1. Summary

Write all Goose SQL migration files for Phase 1 tables, embedded into the binary via `embed.FS`. Migrations run automatically at application startup. The ordering follows FK dependencies: enums → tenants → users → otp_requests → categories → accounts → transactions → audit_logs.

---

## 2. Context & Motivation

Goose is used for schema management because it supports plain SQL files (no Go migration structs required), sequential versioning, and an `embed.FS` workflow that keeps the binary self-contained — no migration files need to be mounted at runtime. The consolidated DDL in `docs/schema.sql` is the source of truth; migration files must match it exactly.

- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.6
- Schema reference: `docs/schema.sql`
- Consumed by: task 1.5.1 (`cmd/api/main.go`) which calls `goose.Up` at boot

---

## 3. Scope

### In scope

- [ ] `internal/platform/db/migrations/00001_create_enums.sql`
- [ ] `internal/platform/db/migrations/00002_create_tenants.sql`
- [ ] `internal/platform/db/migrations/00003_create_users.sql`
- [ ] `internal/platform/db/migrations/00004_create_otp_requests.sql`
- [ ] `internal/platform/db/migrations/00005_create_categories.sql`
- [ ] `internal/platform/db/migrations/00006_create_accounts.sql`
- [ ] `internal/platform/db/migrations/00007_create_transactions.sql`
- [ ] `internal/platform/db/migrations/00008_create_audit_logs.sql`
- [ ] `internal/platform/db/migrations/embed.go` — `//go:embed` directive + `FS` variable
- [ ] Each file includes `-- +goose Up` and `-- +goose Down` sections

### Out of scope

- `master_purchases` table (Phase 2 — task 2.2)
- Seed data / fixture migrations
- Data migrations (no existing data in Phase 1)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                          | Purpose                                    |
| ------ | ------------------------------------------------------------- | ------------------------------------------ |
| CREATE | `internal/platform/db/migrations/00001_create_enums.sql`     | ENUM types: account_type, transaction_type, category_type, user_role, tenant_plan, audit_action |
| CREATE | `internal/platform/db/migrations/00002_create_tenants.sql`   | `tenants` table + index                    |
| CREATE | `internal/platform/db/migrations/00003_create_users.sql`     | `users` table + indexes                    |
| CREATE | `internal/platform/db/migrations/00004_create_otp_requests.sql` | `otp_requests` table + index            |
| CREATE | `internal/platform/db/migrations/00005_create_categories.sql`| `categories` table + index                 |
| CREATE | `internal/platform/db/migrations/00006_create_accounts.sql`  | `accounts` table + indexes                 |
| CREATE | `internal/platform/db/migrations/00007_create_transactions.sql` | `transactions` table + indexes          |
| CREATE | `internal/platform/db/migrations/00008_create_audit_logs.sql`| `audit_logs` table + indexes               |
| CREATE | `internal/platform/db/migrations/embed.go`                   | `//go:embed` FS for binary embedding       |

### Key interfaces / types

```go
// internal/platform/db/migrations/embed.go
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
```

```go
// Usage in cmd/api/main.go (see task 1.5.1)
import (
    "github.com/pressly/goose/v3"
    "github.com/garnizeh/moolah/internal/platform/db/migrations"
)

goose.SetBaseFS(migrations.FS)
if err := goose.Up(db, "."); err != nil {
    log.Fatal(err)
}
```

### Migration file structure

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TYPE account_type AS ENUM ('checking', 'savings', 'credit_card', 'investment');
-- ... (all enums from docs/schema.sql)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS account_type;
-- ...
-- +goose StatementEnd
```

### SQL queries (sqlc)

N/A — migrations are DDL, not DML queries.

### API endpoints (if applicable)

N/A

### Error cases to handle

| Scenario                                  | Handling                                               |
| ----------------------------------------- | ------------------------------------------------------ |
| Migration already applied                 | Goose is idempotent; no action taken                   |
| Migration fails mid-run (partial apply)   | Goose wraps each migration in a transaction; auto-rollback |
| Schema drift (manual change in prod DB)   | Goose version table detects mismatch; startup fails with clear error |

---

## 5. Acceptance Criteria

- [ ] All 8 migration files parse without error via `goose validate`.
- [ ] Running `goose up` on a fresh PostgreSQL database produces an identical schema to `docs/schema.sql`.
- [ ] Running `goose down` to version 0 leaves an empty database (all tables and types removed).
- [ ] `embed.FS` is accessible in tests — migration files are embedded, not read from disk.
- [ ] Integration test: apply all migrations using `testcontainers-go` PostgreSQL instance; assert all tables exist.
- [ ] `docs/ROADMAP.md` row 1.1.6 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                      | Type     | Status     |
| ----------------------------------------------- | -------- | ---------- |
| `github.com/pressly/goose/v3` added to `go.mod` | External | 🔵 backlog |
| `docs/schema.sql` finalised                     | Upstream | ✅ done   |
| Phase 0 complete (module scaffolded)            | Upstream | ✅ done   |

---

## 7. Testing Plan

### Unit tests

N/A — SQL files have no Go logic.

### Integration tests (`//go:build integration`)

- **File:** `internal/platform/db/migrations/migrations_test.go`
- Spin up PostgreSQL via `testcontainers-go`.
- Apply all migrations with `goose.Up`.
- Assert presence of tables: `tenants`, `users`, `otp_requests`, `categories`, `accounts`, `transactions`, `audit_logs`.
- Assert presence of ENUM types: `account_type`, `transaction_type`, `category_type`, `user_role`, `tenant_plan`, `audit_action`.
- Apply `goose.Down` to version 0.
- Assert tables no longer exist.

---

## 8. Open Questions

| # | Question                                                        | Owner | Resolution |
| - | --------------------------------------------------------------- | ----- | ---------- |
| 1 | Use `goose` sequential (00001) or timestamp-based versioning?   | —     | Sequential — more readable, deterministic ordering, no clock-skew issues in team PRs. |
| 2 | Should `-- +goose StatementBegin/End` wrap the whole file or each statement? | — | Wrap each logical block (one per `CREATE TYPE`, one per `CREATE TABLE`). |

---

## 9. Change Log

| Date       | Author | Change                        |
| ---------- | ------ | ----------------------------- |
| 2026-03-07 | —      | Task created from roadmap 1.1.6 |
