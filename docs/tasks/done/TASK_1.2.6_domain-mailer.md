# Task 1.2.6 — Domain Mailer Interface

> **Roadmap Ref:** Phase 1 — MVP › 1.2 Domain Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Define the `Mailer` interface in `internal/domain/mailer.go`. This contract decouples the auth service from any concrete email transport, enabling `SMTPMailer` in production and `NoopMailer` in tests via dependency injection.

---

## 2. Context & Motivation

The auth service (Task 1.4.1) must send OTP emails. By depending on a `domain.Mailer` interface rather than a concrete SMTP implementation, the service remains fully unit-testable with a mock and the transport can be swapped (SMTP → SES → Mailpit) without touching business logic.

---

## 3. Scope

### In scope

- [x] `Mailer` interface with `SendOTP` method.
- [x] Compile-time interface assertion in `smtp_mailer.go`.

### Out of scope

- Concrete `SMTPMailer` implementation (Task 1.1.12).
- `NoopMailer` (Task 1.1.12).
- Future `SendWelcome`, `SendPasswordReset` methods (deferred to Phase 4).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                         | Purpose                  |
| ------ | ---------------------------- | ------------------------ |
| CREATE | `internal/domain/mailer.go`  | `Mailer` interface       |

### Key interfaces / types

```go
// Mailer defines the contract for sending application emails.
// Implementations: SMTPMailer (production), NoopMailer (tests).
type Mailer interface {
    // SendOTP sends a one-time password code to the given email address.
    SendOTP(ctx context.Context, to, code string) error
}
```

### SQL queries (sqlc)

N/A

### API endpoints (if applicable)

N/A

### Error cases to handle

N/A — error handling is the responsibility of concrete implementations.

---

## 5. Acceptance Criteria

- [x] `Mailer` interface is defined in `internal/domain/mailer.go`.
- [x] `SMTPMailer` carries a compile-time guard: `var _ domain.Mailer = (*SMTPMailer)(nil)`.
- [x] `golangci-lint run ./...` passes.
- [x] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency | Type | Status   |
| ---------- | ---- | -------- |
| None       | —    | —        |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

N/A — interface has no logic. Tested via `SMTPMailer` and `NoopMailer` tests.

### Integration tests (`//go:build integration`)

Covered by Task 1.1.13.

---

## 8. Open Questions

| # | Question | Owner | Resolution |
| - | -------- | ----- | ---------- |
| 1 | N/A      | —     | —          |

---

## 9. Change Log

| Date       | Author | Change                         |
| ---------- | ------ | ------------------------------ |
| 2026-03-07 | —      | Task created; already done     |
