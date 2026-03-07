# Task 1.3.7 — Repository: Audit Log

> **Roadmap Ref:** Phase 1 — MVP › 1.3 Repository Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement `AuditRepository` in `internal/platform/repository/audit_repo.go` using the sqlc-generated code. The audit repository is append-only — no update or delete methods exist. It exposes tenant-scoped queries for listing and filtering audit events.

---

## 2. Context & Motivation

The `AuditRepository` interface is defined in `internal/domain/audit.go` (Task 1.2.10). Audit logs provide an immutable compliance trail for all financial and authentication events. Because audit logs are never updated or deleted (in Phase 1), the repository is simpler than other repos — but the `OldValues` / `NewValues` JSON byte slices require careful mapping to avoid nil/empty byte slice inconsistencies across pgx and Go.

---

## 3. Scope

### In scope

- [ ] Concrete `auditRepo` struct implementing `domain.AuditRepository`.
- [ ] Constructor `NewAuditRepository(q *sqlc.Queries) domain.AuditRepository`.
- [ ] Mapping functions between `sqlc.AuditLog` and `domain.AuditLog`.
- [ ] Correct handling of nullable JSON columns (`old_values`, `new_values`).
- [ ] `ListByTenant` applies all `ListAuditLogsParams` filters.

### Out of scope

- Admin `ListAll` (no tenant filter) — that is `AdminAuditRepository` (Task 1.3.8).
- Async audit writing / channels (Phase 5).
- Purge / archival queries (Phase 5).
- HTTP handler (Task 1.5.9).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                          | Purpose                        |
| ------ | --------------------------------------------- | ------------------------------ |
| CREATE | `internal/platform/repository/audit_repo.go` | Concrete AuditRepository impl  |

### Key interfaces / types

```go
type auditRepo struct {
    q *sqlc.Queries
}

func NewAuditRepository(q *sqlc.Queries) domain.AuditRepository {
    return &auditRepo{q: q}
}
```

### SQL queries (sqlc)

All queries already exist in `internal/platform/db/queries/audit_logs.sql` (Task 1.1.7/1.1.8):

| Query name              | sqlc mode | Used by           |
| ----------------------- | --------- | ----------------- |
| `CreateAuditLog`        | `:one`    | `Create`          |
| `ListAuditLogsByTenant` | `:many`   | `ListByTenant`    |
| `ListAuditLogsByEntity` | `:many`   | `ListByEntity`    |

### Error cases to handle

| Scenario              | pgx Error       | Domain Error           |
| --------------------- | --------------- | ---------------------- |
| Row not found         | `pgx.ErrNoRows` | `domain.ErrNotFound`   |

---

## 5. Acceptance Criteria

- [x] All exported types and functions have Go doc comments.
- [x] Struct implements `domain.AuditRepository` (verified by compiler).
- [x] No `Update` or `Delete` methods exist on the struct (append-only enforced).
- [x] `OldValues` / `NewValues` are correctly mapped: DB `NULL` → Go `nil []byte`.
- [x] Every query enforces `tenant_id` isolation.
- [x] `ListByTenant` applies all `ListAuditLogsParams` filters.
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./...` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                        | Type       | Status     |
| --------------------------------- | ---------- | ---------- |
| Task 1.2.10 — `domain/audit.go`   | Upstream   | ✅ done    |
| Task 1.1.7 — sqlc query files     | Upstream   | ✅ done    |
| Task 1.1.8 — sqlc generate        | Upstream   | ✅ done    |
| Task 1.3.9 — integration tests    | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A — repository implementations are tested via integration tests.

### Integration tests (`//go:build integration`)

Covered by Task 1.3.9 — specifically:

- Create audit log entry and verify it is retrievable.
- `ListByEntity` returns correct entries for a given entity.
- `ListByTenant` correctly applies date range and action filters.
- Cross-tenant lookup returns no results.

---

## 8. Open Questions

| # | Question                                                              | Owner | Resolution |
| - | --------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `OldValues` / `NewValues` be stored as `jsonb` or `text` in DB? | —   | `jsonb` for queryability — already defined in migration. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
| 2026-03-07 | Copilot | Task completed             |
