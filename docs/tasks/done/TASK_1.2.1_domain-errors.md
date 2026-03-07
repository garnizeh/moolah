# Task 1.2.1 — Domain Sentinel Errors

> **Roadmap Ref:** Phase 1 — MVP › 1.2 Domain Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Define the canonical set of sentinel errors used throughout the entire codebase. These errors live in `internal/domain/errors.go` and act as the single source of truth for business-logic error conditions, allowing service and handler layers to use `errors.Is`/`errors.As` for clean error-handling and correct HTTP status mapping.

---

## 2. Context & Motivation

Without a shared errors package, each layer invents its own error strings, making consistent HTTP response codes and unit-test assertions impossible. Centralising errors in the domain layer means:

- Services return typed errors; handlers map them to HTTP status codes.
- Unit tests use `errors.Is` instead of brittle string comparisons.
- New error cases are added in one place and propagate everywhere.

---

## 3. Scope

### In scope

- [x] Define all sentinel errors needed by Phase 1 services.
- [x] Add `Error()` string methods where human-readable messages are needed.
- [x] Write unit tests confirming `errors.Is` works for each sentinel.

### Out of scope

- HTTP-level error response types (deferred to handler layer in 1.5.x).
- gRPC status code mapping (deferred to Phase 5).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                           | Purpose                          |
| ------ | ------------------------------ | -------------------------------- |
| CREATE | `internal/domain/errors.go`    | All sentinel error declarations  |
| CREATE | `internal/domain/errors_test.go` | Unit tests for error wrapping   |

### Key interfaces / types

```go
// Sentinel errors — all in internal/domain/errors.go
var (
    // ErrNotFound is returned when a requested resource does not exist or has been soft-deleted.
    ErrNotFound = errors.New("not found")

    // ErrForbidden is returned when the caller lacks permission for a resource.
    ErrForbidden = errors.New("forbidden")

    // ErrConflict is returned when an operation would violate a uniqueness constraint.
    ErrConflict = errors.New("conflict")

    // ErrInvalidInput is returned when request validation fails.
    ErrInvalidInput = errors.New("invalid input")

    // ErrInvalidOTP is returned when an OTP code is wrong, expired, or already used.
    ErrInvalidOTP = errors.New("invalid or expired OTP")

    // ErrOTPRateLimited is returned when too many OTP requests are made.
    ErrOTPRateLimited = errors.New("OTP rate limit exceeded")

    // ErrUnauthorized is returned when no valid token is present.
    ErrUnauthorized = errors.New("unauthorized")

    // ErrTokenExpired is returned when the PASETO token has expired.
    ErrTokenExpired = errors.New("token expired")
)
```

### SQL queries (sqlc)

N/A — this task contains no database interaction.

### API endpoints (if applicable)

N/A — this task contains no HTTP handlers.

### Error cases to handle

N/A — this task *defines* the error cases used by others.

---

## 5. Acceptance Criteria

- [ ] All sentinel errors are defined in `internal/domain/errors.go`.
- [ ] `errors.Is(fmt.Errorf("wrapped: %w", domain.ErrNotFound), domain.ErrNotFound)` returns `true`.
- [ ] Unit tests cover every sentinel.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency       | Type     | Status  |
| ---------------- | -------- | ------- |
| None             | —        | —       |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/domain/errors_test.go`
- **Cases:**
  - Each sentinel unwraps correctly via `errors.Is`.
  - Wrapped errors (using `fmt.Errorf("%w", ...)`) still match via `errors.Is`.

### Integration tests (`//go:build integration`)

N/A

---

## 8. Open Questions

| # | Question | Owner | Resolution |
| - | -------- | ----- | ---------- |
| 1 | Should we add a `ErrPlanQuotaExceeded` now or defer to Phase 4? | — | — |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
