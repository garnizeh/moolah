# Task 1.5.4 — `handler/auth_handler.go` — `RequestOTP`, `VerifyOTP`, `RefreshToken`

> **Roadmap Ref:** Phase 1 — MVP › 1.5 HTTP Handler Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-09
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the authentication HTTP handler in `internal/handler/auth_handler.go`. It exposes three endpoints: request an OTP code, verify an OTP code and receive a PASETO token, and refresh an existing token. This is the gateway for all user sessions in the system.

---

## 2. Context & Motivation

The `AuthService` is fully implemented (Task 1.4.1) but has no HTTP entry point. This task creates the handler that translates HTTP request/response to service calls, enforces input validation, maps domain errors to HTTP status codes, and writes structured JSON responses. See `docs/ARCHITECTURE.md` and roadmap item 1.5.4.

---

## 3. Scope

### In scope

- [x] `internal/handler/auth_handler.go` — `AuthHandler` struct + 3 HTTP handler methods.
- [x] `POST /v1/auth/otp/request` — validate email, call `AuthService.RequestOTP`.
- [x] `POST /v1/auth/otp/verify` — validate email + code, call `AuthService.VerifyOTP`, return token.
- [x] `POST /v1/auth/token/refresh` — extract bearer token, call `AuthService.RefreshToken`, return new token.
- [x] Request struct validation via `go-playground/validator`.
- [x] Domain error → HTTP status mapping.
- [x] Unit tests in `internal/handler/auth_handler_test.go` with mocked `AuthService`.

### Out of scope

- Rate-limit middleware — already implemented (Task 1.1.10); applied in `routes.go`.
- OTP delivery (email) — handled inside `AuthService`.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                   | Purpose                              |
| ------ | -------------------------------------- | ------------------------------------ |
| CREATE | `internal/handler/auth_handler.go`     | HTTP handler for auth endpoints      |
| CREATE | `internal/handler/auth_handler_test.go`| Unit tests with mocked AuthService   |

### Request / Response types

```go
type RequestOTPRequest struct {
    Email string `json:"email" validate:"required,email"`
}

type VerifyOTPRequest struct {
    Email string `json:"email" validate:"required,email"`
    Code  string `json:"code"  validate:"required,len=6"`
}

type TokenResponse struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
}
```

### API endpoints

| Method | Path                       | Auth Required | Description                    |
| ------ | -------------------------- | ------------- | ------------------------------ |
| POST   | `/v1/auth/otp/request`     | ❌ None       | Send OTP to email              |
| POST   | `/v1/auth/otp/verify`      | ❌ None       | Verify OTP, receive token      |
| POST   | `/v1/auth/token/refresh`   | ✅ Bearer     | Refresh session token          |

### Error cases to handle

| Scenario                    | Sentinel Error              | HTTP Status |
| --------------------------- | --------------------------- | ----------- |
| Invalid request body        | —                           | `400`       |
| Validation failure          | —                           | `422`       |
| OTP rate limited            | `domain.ErrOTPRateLimited`  | `429`       |
| OTP invalid / expired       | `domain.ErrInvalidOTP`      | `401`       |
| Token expired               | `domain.ErrTokenExpired`    | `401`       |
| Unauthorized                | `domain.ErrUnauthorized`    | `401`       |
| Not found                   | `domain.ErrNotFound`        | `404`       |

---

## 5. Acceptance Criteria

- [x] All 3 endpoints decode, validate, and call the service correctly.
- [x] All domain error sentinels map to the correct HTTP status codes.
- [x] JSON responses follow a consistent envelope: `{"data": ...}` for success and `{"error": "..."}` for failures.
- [x] Unit tests cover happy paths and all named error cases.
- [x] Test coverage for handler ≥ 80%.
- [x] `golangci-lint run ./...` passes with zero issues.
- [x] `gosec ./......` passes with zero issues.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                          | Type     | Status  |
| ----------------------------------- | -------- | ------- |
| Task 1.4.1 — `service/auth_service` | Upstream | ✅ done |
| Task 1.1.9 — Auth middleware        | Upstream | ✅ done |
| `domain.AuthService` interface      | Upstream | ✅ done |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/handler/auth_handler_test.go`
- **Cases:**
  - `RequestOTP`: valid request → `202 Accepted`.
  - `RequestOTP`: invalid email → `422`.
  - `RequestOTP`: rate-limited → `429`.
  - `VerifyOTP`: valid code → `200 OK` with `TokenResponse`.
  - `VerifyOTP`: wrong code → `401`.
  - `RefreshToken`: valid token → `200 OK` with new `TokenResponse`.
  - `RefreshToken`: expired token → `401`.

---

## 8. Open Questions

| # | Question                                             | Owner | Resolution |
| - | ---------------------------------------------------- | ----- | ---------- |
| 1 | Should `RequestOTP` return `200` or `202`?           | —     | `202 Accepted` — the OTP is queued for delivery. |
| 2 | Should the response envelope use `{"data": ...}` or flat JSON? | — | Flat for auth responses; `{"data": ...}` for resource responses. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
| 2026-03-09 | —      | Completed handler implementation and tests |
