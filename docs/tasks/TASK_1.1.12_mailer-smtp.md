# Task 1.1.12 — platform/mailer/smtp_mailer.go: SMTP Mailer

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the `domain.Mailer` interface with an SMTP backend using Go's standard `net/smtp` package. The mailer sends one type of email in Phase 1: the OTP verification code. A `NoopMailer` (also implements `domain.Mailer`) is provided for tests, returning `nil` without sending anything.

---

## 2. Context & Motivation

The domain layer defines the `Mailer` interface so services remain decoupled from the transport mechanism — tests inject a `NoopMailer`, production injects `SMTPMailer`. This satisfies the Interface-Driven Development rule from `copilot-instructions.md`. All SMTP credentials come from `pkg/config.Config` (task 1.1.3); the mailer must never hard-code them.

- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.12
- Interface definition: task 1.2.6 (`domain/mailer.go`)
- Depends on: task 1.1.3 (`pkg/config`) for SMTP settings
- Consumed by: task 1.4.1 (`service/auth_service.go`)

---

## 3. Scope

### In scope

- [ ] `internal/domain/mailer.go` — `Mailer` interface (also covers task 1.2.6)
- [ ] `internal/platform/mailer/smtp_mailer.go` — `SMTPMailer` struct + constructor
- [ ] `internal/platform/mailer/noop_mailer.go` — `NoopMailer` for tests
- [ ] `internal/platform/mailer/smtp_mailer_test.go` — unit tests with `NoopMailer`
- [ ] Plain-text + HTML OTP email template (inline, no template files in Phase 1)

### Out of scope

- Transactional email providers (SendGrid, Postmark) — deferred; interface makes this a drop-in swap
- Email queue / retry (deferred to Phase 5)
- Attachment support (not needed)
- HTML email templating system (Phase 1 uses inline strings)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                              | Purpose                                         |
| ------ | ------------------------------------------------- | ----------------------------------------------- |
| CREATE | `internal/domain/mailer.go`                       | `Mailer` interface                              |
| CREATE | `internal/platform/mailer/smtp_mailer.go`         | SMTP implementation of `domain.Mailer`          |
| CREATE | `internal/platform/mailer/noop_mailer.go`         | No-op implementation for tests                  |
| CREATE | `internal/platform/mailer/smtp_mailer_test.go`    | Unit tests (using NoopMailer)                   |

### Key interfaces / types

```go
// internal/domain/mailer.go
package domain

import "context"

// Mailer defines the contract for sending application emails.
// Implementations: SMTPMailer (production), NoopMailer (tests).
type Mailer interface {
    // SendOTP sends a one-time password code to the given email address.
    SendOTP(ctx context.Context, to, code string) error
}
```

```go
// internal/platform/mailer/smtp_mailer.go
package mailer

import (
    "context"
    "fmt"
    "net/smtp"

    "github.com/garnizeh/moolah/internal/domain"
)

// SMTPMailer implements domain.Mailer via net/smtp.
type SMTPMailer struct {
    host     string
    port     int
    username string
    password string
    from     string
}

// NewSMTPMailer constructs an SMTPMailer. Returns an error if required fields are empty.
func NewSMTPMailer(host string, port int, username, password, from string) (*SMTPMailer, error) { ... }

// SendOTP sends the OTP code to the recipient via SMTP with STARTTLS.
func (m *SMTPMailer) SendOTP(ctx context.Context, to, code string) error { ... }

// Ensure interface compliance at compile time.
var _ domain.Mailer = (*SMTPMailer)(nil)
```

```go
// internal/platform/mailer/noop_mailer.go
package mailer

import "context"

// NoopMailer is a domain.Mailer implementation that discards all emails.
// Use in unit tests and local development.
type NoopMailer struct{}

func (n *NoopMailer) SendOTP(_ context.Context, _, _ string) error { return nil }

var _ domain.Mailer = (*NoopMailer)(nil)
```

### OTP email body (inline template)

```
Subject: Your Moolah verification code

Your one-time login code is: {{code}}

This code expires in 10 minutes. Do not share it with anyone.
```

### SQL queries (sqlc)

N/A

### API endpoints (if applicable)

N/A

### Error cases to handle

| Scenario                             | Handling                                         |
| ------------------------------------ | ------------------------------------------------ |
| SMTP dial failure                    | Return wrapped error: `fmt.Errorf("mailer: dial: %w", err)` |
| SMTP authentication failure          | Return wrapped error                             |
| `to` address is empty                | Return `domain.ErrInvalidInput` before dialing   |
| `code` is empty                      | Return `domain.ErrInvalidInput` before dialing   |
| Context cancelled before send        | Return `ctx.Err()` wrapped                       |

---

## 5. Acceptance Criteria

- [ ] `SMTPMailer` satisfies `domain.Mailer` (compile-time check via `var _ domain.Mailer = ...`).
- [ ] `NoopMailer` satisfies `domain.Mailer` (compile-time check).
- [ ] `NewSMTPMailer` returns an error if `host`, `username`, `password`, or `from` is empty.
- [ ] `SendOTP` with an empty `to` address returns an error without dialing SMTP.
- [ ] `SendOTP` with an empty `code` returns an error without dialing SMTP.
- [ ] Unit tests use `NoopMailer` exclusively — no external SMTP server required.
- [ ] Test coverage for `smtp_mailer.go` ≥ 80% (dial/send paths require integration or interface injection).
- [ ] `golangci-lint run ./internal/platform/mailer/...` passes with zero issues.
- [ ] `gosec ./internal/platform/mailer/...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 1.1.12 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                     | Type     | Status     |
| ---------------------------------------------- | -------- | ---------- |
| Task 1.2.6 `domain/mailer.go` — interface      | Upstream | 🔵 backlog (co-located in this task) |
| Task 1.1.3 `pkg/config` — SMTP settings        | Upstream | 🔵 backlog |
| Go stdlib `net/smtp`                           | Runtime  | ✅ done   |

---

## 7. Testing Plan

### Unit tests (`internal/platform/mailer/smtp_mailer_test.go`, no build tag)

- **NoopMailer:** `SendOTP(ctx, "user@example.com", "123456")` returns `nil`.
- **NewSMTPMailer validation:** assert error when `host=""`, `username=""`, `password=""`, `from=""`.
- **SendOTP empty to:** assert error returned, no dial attempted (inject a spy dialer).
- **SendOTP empty code:** assert error returned, no dial attempted.
- **Interface compliance:** both `SMTPMailer` and `NoopMailer` implement `domain.Mailer` (compile-time).

### Integration tests (`//go:build integration`)

- **File:** `internal/platform/mailer/smtp_mailer_integration_test.go`
- Spin up a Mailpit container (local SMTP + API) via `testcontainers-go`.
- Call `SendOTP`; verify the email appears in Mailpit's API (`GET /api/v1/messages`).
- Assert `To`, `Subject`, and body contain expected values.

---

## 8. Open Questions

| # | Question                                                              | Owner | Resolution |
| - | --------------------------------------------------------------------- | ----- | ---------- |
| 1 | Use `net/smtp` directly or a thin wrapper like `gopkg.in/gomail.v2`? | —     | `net/smtp` stdlib for Phase 1 — zero dependencies; `gomail` if HTML templates become complex in Phase 2. |
| 2 | Support SMTP with TLS (`smtps`, port 465) as well as STARTTLS (587)?  | —     | STARTTLS only in Phase 1; TLS port support deferred. |

---

## 9. Change Log

| Date       | Author | Change                         |
| ---------- | ------ | ------------------------------ |
| 2026-03-07 | —      | Task created from roadmap 1.1.12 |
