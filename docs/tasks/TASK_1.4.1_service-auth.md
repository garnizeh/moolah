# Task 1.4.1 — `service/auth_service.go` + unit tests

> **Roadmap Ref:** Phase 1 — MVP › 1.4 Service Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** L

---

## 1. Summary

Implement the `AuthService` which orchestrates the OTP-based, passwordless authentication flow. It coordinates the `AuthRepository`, `UserRepository`, `Mailer`, and `pkg/paseto` to deliver `RequestOTP`, `VerifyOTP`, and `RefreshToken` use cases. All operations are fully unit-tested with mocked dependencies.

---

## 2. Context & Motivation

The system uses email + OTP as the sole authentication strategy (no passwords). See `docs/ARCHITECTURE.md` and roadmap item 1.4.1. The `AuthRepository` and `UserRepository` (Tasks 1.3.3 and 1.3.2) provide the persistence layer. The `AuthService` sits between these repositories and the HTTP handler layer, enforcing all business rules:

- Rate limiting hints (service returns structured errors for the handler to act on).
- OTP expiry and single-use enforcement.
- First-login auto-provisioning: if the user does not exist yet, they are not created at OTP request time; the user must already exist (invited by a tenant admin).
- JWT (PASETO) token generation on successful verification.
- Audit trail creation for every auth event.

---

## 3. Scope

### In scope

- [✅] `internal/service/auth_service.go` — concrete `AuthService` struct with constructor.
- [✅] `internal/domain/auth.go` — add `AuthService` interface definition.
- [✅] `RequestOTP(ctx, email string) error` — validate email, look up user, create OTP, send via Mailer, create audit log (`otp_requested`).
- [✅] `VerifyOTP(ctx, email, code string) (*TokenPair, error)` — retrieve active OTP, compare bcrypt hash, mark used, update last login, create audit log (`otp_verified` or `login_failed`), return signed PASETO token pair.
- [✅] `RefreshToken(ctx, refreshToken string) (*TokenPair, error)` — validate refresh token, reissue access token.
- [✅] `TokenPair` value type: `{ AccessToken string; RefreshToken string; ExpiresAt time.Time }`.
- [✅] Full unit tests with gomock/testify in `internal/service/auth_service_test.go`.

### Out of scope

- HTTP handler (`handler/auth_handler.go`) — Task 1.5.4.
- Rate-limiting logic — already in middleware (Task 1.1.10); service only returns errors.
- `DeleteExpiredOTPRequests` — background cleanup job (future task).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                    | Purpose                                            |
| ------ | --------------------------------------- | -------------------------------------------------- |
| MODIFY | `internal/domain/auth.go`               | Add `AuthService` interface + `TokenPair` type     |
| CREATE | `internal/service/auth_service.go`      | Concrete service implementation                    |
| CREATE | `internal/service/auth_service_test.go` | Unit tests with mocked deps                        |

### Key interfaces / types

```go
// TokenPair holds both tokens returned after successful OTP verification.
type TokenPair struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    ExpiresAt    time.Time `json:"expires_at"`
}

// AuthService defines the business-logic contract for the OTP auth flow.
type AuthService interface {
    // RequestOTP validates the email, generates an OTP, persists it, and mails
    // the code to the user. Returns ErrNotFound if the user does not exist.
    RequestOTP(ctx context.Context, email string) error

    // VerifyOTP validates the OTP code for the given email. On success it marks
    // the OTP as used, updates the user's last-login timestamp, records an audit
    // log, and returns a fresh PASETO token pair.
    VerifyOTP(ctx context.Context, email, code string) (*TokenPair, error)

    // RefreshToken validates an existing refresh token and returns a new token
    // pair with a refreshed expiry.
    RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
}
```

### Business rules

1. `RequestOTP`:
   - Look up user by email via `UserRepository.GetByEmail`. If not found, return `domain.ErrNotFound`.
   - Generate a 6-digit OTP via `pkg/otp.Generate()`.
   - Hash the code via `pkg/otp.Hash(code)`.
   - Persist via `AuthRepository.CreateOTPRequest` with `ExpiresAt = now + 10 min`.
   - Send email via `Mailer.SendOTP(ctx, email, code)`.
   - Append audit log: action `otp_requested`, actor = user ID, entity = `otp_request`.

2. `VerifyOTP`:
   - Retrieve active OTP via `AuthRepository.GetActiveOTPRequest(ctx, email)`. On `ErrInvalidOTP`, append audit log `login_failed` and return `ErrInvalidOTP`.
   - Compare plain code against stored hash via `pkg/otp.Verify(code, hash)`. On mismatch, append audit log `login_failed` and return `ErrInvalidOTP`.
   - Mark OTP used via `AuthRepository.MarkOTPUsed`.
   - Update last login via `UserRepository.UpdateLastLogin`.
