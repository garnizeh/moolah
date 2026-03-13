# Task 2.3 — sqlc Queries: `master_purchases`

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › Infrastructure
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-12
> **Assignee:** GitHub Copilot
> **Estimated Effort:** S

---

## 1. Summary

Write the raw SQL query file `internal/platform/db/queries/master_purchases.sql` for all `MasterPurchaseRepository` operations and run `sqlc generate` to produce the corresponding Go code. Follows the same pattern as Phase 1 query files.

---

## 2. Context & Motivation

`sqlc` generates type-safe Go code from annotated SQL. Every query must include `WHERE tenant_id = $1` and `AND deleted_at IS NULL` to enforce multi-tenancy and soft-delete semantics (project rule). The generated `Querier` interface and structs must be usable by the concrete repository (Task 2.4) and mockable in unit tests via the centralized mocks (`internal/testutil/mocks`).

---

## 3. Scope

### In scope

- [x] SQL file with all required named queries (see §4).
- [x] `sqlc generate` runs without errors.
- [x] Generated code committed to `internal/platform/db/sqlc/` (or equivalent generated output path).
- [x] `MockQuerier` in `internal/testutil/mocks` updated to include new methods.

### Out of scope

- Concrete repository implementation (Task 2.4).
- Business logic (Task 2.5).
- Goose migration (Task 2.2).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                      | Purpose                            |
| ------ | --------------------------------------------------------- | ---------------------------------- |
| CREATE | `internal/platform/db/queries/master_purchases.sql`       | Named sqlc queries                 |
| MODIFY | `internal/testutil/mocks/querier.go`                    | Add mock methods for new queries   |

### SQL queries (sqlc)

Corrected SQL queries implemented in `internal/platform/db/queries/master_purchases.sql`.

---

## 5. Acceptance Criteria

- [x] All queries include `WHERE tenant_id = $1`.
- [x] All read queries include `AND deleted_at IS NULL`.
- [x] `sqlc generate` succeeds with zero errors or warnings.
- [x] Generated Go types map `total_amount_cents` and `paid_installments` to `int64`/`int32` (no float).
- [x] `MockQuerier` includes stub methods for all 8 new queries.
- [x] `golangci-lint run ./...` passes on generated code.
- [x] CI `sqlc-diff` check passes (no diff between SQL and generated code).
- [x] `docs/ROADMAP.md` row 2.3 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                   | Type     | Status       |
| -------------------------------------------- | -------- | ------------ |
| Task 2.1 — `domain/master_purchase.go`       | Upstream | ✅ done      |
| Task 2.2 — Goose migration (table exists)    | Upstream | ✅ done      |
| Task 2.4 — Repository impl (consumer)        | Downstream | ✅ done    |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

Verified via mock build and repository unit tests.

### Integration tests (`//go:build integration`)

Covered by Task 2.4 repository integration tests (Pending integration run).

---

## 8. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
| 2026-03-12 | GitHub Copilot | Task completed: SQL generated and mocks updated |
