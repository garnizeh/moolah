# Task 1.2.5 — Domain Auth Entity & Repository Interface

> **Roadmap Ref:** Phase 1 — MVP › 1.2 Domain Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Define the `OTPRequest` domain entity and `AuthRepository` interface in `internal/domain/auth.go`. The `OTPRequest` tracks an in-flight OTP challenge (email, hashed code, TTL, used flag). The `AuthRepository` allows the auth service to be fully tested without touching the database.

---

## 2. Context & Motivation

The authentication flow (Task 1.4.1) is OTP-only. When a user requests a login code, an `OTPRequest` row is created; when they verify the code, the row is marked `used=true`. Sitting in the domain layer, `AuthRepository` decouples the auth service from pgx/sqlc implementation details and makes the OTP verification logic fully unit-testable via mocks.

---

## 3. Scope

### In scope

- [x] `OTPRequest` struct.
- [x] `CreateOTPRequestInput` value object.
- [x] `AuthRepository` interface: `CreateOTPRequest`, `GetActiveOTPRequest`, `MarkOTPUsed`.
- [x] `Claims` struct for PASETO token payload (user_id, tenant_id, role, expiry).

### Out of scope

- Token issuance/verification logic (lives in `pkg/paseto`).
- Concrete repository implementation (Task 1.3.3).
- Auth service orchestration (Task 1.4.1).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                      | Purpose                                      |
| ------ | ------------------------- | -------------------------------------------- |
| CREATE | `internal/domain/auth.go` | OTPRequest entity, input types, repo interface, Claims struct |

### Key interfaces / types

```go
// OTPRequest represents a pending or consumed OTP challenge.
type OTPRequest struct {
 ExpiresAt time.Time `json:"expires_at"`
 CreatedAt time.Time `json:"created_at"`
 ID        string    `json:"id"`
 Email     string    `json:"email"`
 CodeHash  string    `json:"-"` // bcrypt hash of the 6-digit code, never serialized
 Used      bool      `json:"used"`
}

type CreateOTPRequestInput struct {
 ExpiresAt time.Time `validate:"required"`
 Email     string    `validate:"required,email"`
 CodeHash  string    `validate:"required"`
}

// AuthRepository defines persistence operations for OTP challenges.
type AuthRepository interface {
    // CreateOTPRequest persists a new OTP challenge.
    CreateOTPRequest(ctx context.Context, input CreateOTPRequestInput) (*OTPRequest, error)
    // GetActiveOTPRequest retrieves the most recent unused, non-expired OTP for the given email.
    GetActiveOTPRequest(ctx context.Context, email string) (*OTPRequest, error)
    // MarkOTPUsed marks the given OTP request as consumed.
    MarkOTPUsed(ctx context.Context, id string) error
    // DeleteExpiredOTPRequests removes all expired OTP rows (called by a periodic cleanup job).
    DeleteExpiredOTPRequests(ctx context.Context) error
}

// Claims holds the data encoded in a PASETO token.
type Claims struct {
 IssuedAt  time.Time `json:"issued_at"`
 ExpiresAt time.Time `json:"expires_at"`
 UserID    string    `json:"user_id"`
 TenantID  string    `json:"tenant_id"`
 Role      Role      `json:"role"`
}
```

### SQL queries (sqlc)

Queries already generated in Task 1.1.7/1.1.8 under `internal/platform/db/queries/auth.sql`.

### API endpoints (if applicable)

N/A — endpoints are registered in Task 1.5.4.

### Error cases to handle

| Scenario                    | Sentinel Error            | HTTP Status |
| --------------------------- | ------------------------- | ----------- |
| No active OTP found         | `domain.ErrInvalidOTP`    | `401`       |
| OTP already used / expired  | `domain.ErrInvalidOTP`    | `401`       |
| User not found for email    | `domain.ErrNotFound`      | `404`       |

---

## 5. Acceptance Criteria

- [x] All exported types and functions have Go doc comments.
- [x] `AuthRepository` interface is defined in `internal/domain/auth.go`.
- [x] `OTPRequest` does not store the plaintext code — only the `bcrypt` hash.
- [x] `Claims` struct is usable by `pkg/paseto` without importing domain (or vice versa — decide direction).
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./...` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                       | Type     | Status     |
| -------------------------------- | -------- | ---------- |
| Task 1.2.1 — `domain/errors.go`  | Upstream | 🔵 backlog |
| Task 1.2.2 — `domain/role.go`    | Upstream | ✅ done    |
| Task 1.1.5 — `pkg/otp`           | Upstream | ✅ done    |
| Task 1.1.4 — `pkg/paseto`        | Upstream | ✅ done    |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A (pure types/interfaces — tested via auth service in 1.4.1)

### Integration tests (`//go:build integration`)

Covered by Task 1.3.3 repository integration tests.

---

## 8. Open Questions

| # | Question                                                                        | Owner | Resolution |
| - | ------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `Claims` live in `domain` or `pkg/paseto`? (circular import risk)         | —     | Lives in domain. |
| 2 | Should `DeleteExpiredOTPRequests` be triggered by a background goroutine or cron? | —     | Cleanup job. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
| 2026-03-07 | —      | Task completed |
