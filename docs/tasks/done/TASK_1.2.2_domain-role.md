# Task 1.2.2 — Domain Role Type

> **Roadmap Ref:** Phase 1 — MVP › 1.2 Domain Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Define the `Role` type, constants (`sysadmin`, `admin`, `member`), and helper methods (`Level()`, `CanAccess()`) used throughout the entire codebase for RBAC checks in middleware and service layers.

---

## 2. Context & Motivation

A central `Role` type prevents string-based role comparisons scattered across the codebase. The `Level()` numeric weight and `CanAccess()` helper enable clean, hierarchical permission checks in both `RequireRole` middleware and service-layer guards without repetitive `switch` blocks.

---

## 3. Scope

### In scope

- [x] `Role` string type with `const` values.
- [x] `Level() int` method for numeric comparison.
- [x] `CanAccess(target Role) bool` method for hierarchical RBAC.
- [x] Unit tests covering all role levels and cross-role access checks.

### Out of scope

- Permission lists / ACL (future requirement).
- Persisting roles in the DB (handled by `UserRole` enum in sqlc models).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                        | Purpose                         |
| ------ | --------------------------- | ------------------------------- |
| CREATE | `internal/domain/role.go`   | `Role` type + constants + methods |
| CREATE | `internal/domain/role_test.go` | Unit tests                   |

### Key interfaces / types

```go
type Role string

const (
    RoleSysadmin Role = "sysadmin"
    RoleAdmin    Role = "admin"
    RoleMember   Role = "member"
)

func (r Role) Level() int { ... }
func (r Role) CanAccess(target Role) bool { return r.Level() >= target.Level() }
```

### SQL queries (sqlc)

N/A

### API endpoints (if applicable)

N/A

### Error cases to handle

N/A

---

## 5. Acceptance Criteria

- [x] All exported types and functions have Go doc comments.
- [x] `CanAccess` correctly enforces hierarchy: `sysadmin >= admin >= member`.
- [x] Unit tests cover all levels and access checks.
- [x] `golangci-lint run ./...` passes.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency | Type | Status     |
| ---------- | ---- | ---------- |
| None       | —    | —          |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/domain/role_test.go`
- **Cases:**
  - `Level()` returns correct int for each role.
  - `CanAccess()` returns `true` for same or higher role.
  - `CanAccess()` returns `false` for lower role.
  - Unknown role has level 0 and cannot access anything.

### Integration tests (`//go:build integration`)

N/A

---

## 8. Open Questions

| # | Question | Owner | Resolution |
| - | -------- | ----- | ---------- |
| 1 | N/A      | —     | —          |

---

## 9. Change Log

| Date       | Author | Change                       |
| ---------- | ------ | ---------------------------- |
| 2026-03-07 | —      | Task created; already done   |
