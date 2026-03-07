# Task 1.3.3 — Repository: Auth (OTP)

> **Roadmap Ref:** Phase 1 — MVP › 1.3 Repository Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement `AuthRepository` in `internal/platform/repository/auth_repo.go` using the sqlc-generated code. This repository manages the lifecycle of OTP challenges: creation, active lookup, marking as used, and cleanup of expired rows.

---

## 2. Context & Motivation

The `AuthRepository` interface is defined in `internal/domain/auth.go` (Task 1.2.5). OTP challenges are short-lived records (10 min TTL) that power the passwordless login flow. The repository must enforce business rules at the query level: `GetActiveOTPRequest` returns only the most recent unused, non-expired row, ensuring replay protection without service-layer round-trips.

---

## 3. Scope

### In scope

- [ ] Concrete `authRepo` struct implementing `domain.AuthRepository`.
- [ ] Constructor `NewAuthRepository(q *sqlc.Queries) domain.AuthRepository`.
- [ ] Mapping functions between `sqlc.OtpRequest` and `domain.OTPRequest`.
- [ ] Error translation: `pgx.ErrNoRows` on active OTP lookup → `domain.ErrInvalidOTP`.

### Out of scope

- OTP code generation (lives in `pkg/otp`).
- Rate limiting (lives in `platform/middleware/ratelimit.go`).
- Auth service orchestration (Task 1.4.1).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                         | Purpose                        |
| ------ | -------------------------------------------- | ------------------------------ |
| CREATE | `internal/platform/repository/auth_repo.go` | Concrete AuthRepository impl   |

### Key interfaces / types

```go
type authRepo struct {
    q *sqlc.Queries
}

func NewAuthRepository(q *sqlc.Queries) domain.AuthRepository {
    return &authRepo{q: q}
}
```

### SQL queries (sqlc)

All queries already exist in `internal/platform/db/queries/auth.sql` (Task 1.1.7/1.1.8):

| Query name                    | sqlc mode | Used by                      |
| ----------------------------- | --------- | ---------------------------- |
| `CreateOTPRequest`            | `:one`    | `CreateOTPRequest`           |
| `GetActiveOTPRequest`         | `:one`    | `GetActiveOTPRequest`        |
| `MarkOTPUsed`                 | `:exec`   | `MarkOTPUsed`                |
| `DeleteExpiredOTPRequests`    | `:exec`   | `DeleteExpiredOTPRequests`   |

### Error cases to handle

| Scenario                    | pgx Error       | Domain Error            |
| --------------------------- | --------------- | ----------------------- |
| No active OTP found         | `pgx.ErrNoRows` | `domain.ErrInvalidOTP`  |

---

## 5. Acceptance Criteria

- [ ] All exported types and functions have Go doc comments.
- [ ] Struct implements `domain.AuthRepository` (verified by compiler).
- [ ] `GetActiveOTPRequest` returns `domain.ErrInvalidOTP` (not `ErrNotFound`) when no active OTP exists.
- [ ] `DeleteExpiredOTPRequests` deletes all rows where `expires_at < NOW() OR used = true`.
- [ ] All pgx errors are translated to domain sentinel errors.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                      | Type       | Status     |
| ------------------------------- | ---------- | ---------- |
| Task 1.2.5 — `domain/auth.go`   | Upstream   | ✅ done    |
| Task 1.1.7 — sqlc query files   | Upstream   | ✅ done    |
| Task 1.1.8 — sqlc generate      | Upstream   | ✅ done    |
| Task 1.3.9 — integration tests  | Downstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A — repository implementations are tested via integration tests.

### Integration tests (`//go:build integration`)

Covered by Task 1.3.9 — specifically:

- Create OTP and verify it is retrievable.
- Verify expired OTP is not returned by `GetActiveOTPRequest`.
- Verify used OTP is not returned by `GetActiveOTPRequest`.
- Verify `DeleteExpiredOTPRequests` removes correct rows.

---

## 8. Open Questions

| # | Question                                                                     | Owner | Resolution |
| - | ---------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should we enforce a max of one active OTP per email at the DB or app layer?  | —     | App layer in the auth service — the query returns the most recent; older ones are ignored. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
