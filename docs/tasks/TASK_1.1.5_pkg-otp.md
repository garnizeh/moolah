# Task 1.1.5 — pkg/otp: Cryptographically Secure 6-Digit OTP

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-07
> **Assignee:** GitHub Copilot
> **Estimated Effort:** S

---

## 1. Summary

Implement a `pkg/otp` package that generates a cryptographically secure 6-digit one-time password and returns its bcrypt hash for storage. A companion `Verify` function checks a plain-text candidate against the stored hash. The plain-text code is never persisted — only the hash reaches the database (`otp_requests.code_hash`).

---

## 2. Context & Motivation

Email + OTP is the sole authentication strategy (no passwords). The 6-digit code must be generated from `crypto/rand` to avoid predictability. bcrypt hashing before storage ensures that a database leak does not expose valid codes. The 10-minute TTL enforcement is done at the service layer using `otp_requests.expires_at`; this package handles only generation and verification.

- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.5
- Schema reference: `docs/schema.sql` → `otp_requests.code_hash TEXT NOT NULL`
- Consumed by: task 1.4.1 (`service/auth_service.go`)

---

## 3. Scope

### In scope

- [x] `pkg/otp/otp.go` — `Generate() (plain, hash string, err error)` and `Verify(plain, hash string) bool`
- [x] `pkg/otp/otp_test.go` — generation uniqueness, format, and verify correctness tests
- [x] bcrypt cost constant = 12 (good default for 2026 hardware; adjustable via a package-level constant)

### Out of scope

- TTL enforcement (handled in `service/auth_service.go`)
- Rate limiting (handled in task 1.1.10 `middleware/ratelimit.go`)
- Code delivery (handled in task 1.1.12 `mailer/smtp_mailer.go`)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                         | Purpose                              |
| ------ | ---------------------------- | ------------------------------------ |
| CREATE | `pkg/otp/otp.go`             | Generate + Verify functions          |
| CREATE | `pkg/otp/otp_test.go`        | Unit tests                           |

### Key interfaces / types

```go
// pkg/otp/otp.go
package otp

import (
    "crypto/rand"
    "fmt"
    "math/big"

    "golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// Generate produces a cryptographically random 6-digit code and its bcrypt hash.
// The caller must send `plain` to the user and persist `hash` in the database.
// The plain-text code must never be stored.
func Generate() (plain, hash string, err error) {
    const digits = 6
    max := big.NewInt(1_000_000)
    n, err := rand.Int(rand.Reader, max)
    if err != nil {
        return "", "", fmt.Errorf("otp: generate: %w", err)
    }
    plain = fmt.Sprintf("%06d", n.Int64())
    b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
    if err != nil {
        return "", "", fmt.Errorf("otp: hash: %w", err)
    }
    return plain, string(b), nil
}

// Verify returns true if plain matches the stored bcrypt hash.
func Verify(plain, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
```

### SQL queries (sqlc)

N/A — hash is passed as a parameter to `platform/repository/auth_repo.go`.

### API endpoints (if applicable)

N/A

### Error cases to handle

| Scenario                         | Handling                                        |
| -------------------------------- | ----------------------------------------------- |
| `crypto/rand` read failure       | Return wrapped error; caller propagates to HTTP 500 |
| `bcrypt.GenerateFromPassword` fail | Return wrapped error                          |
| Wrong plain code in `Verify`     | Return `false` (no error — not exceptional)     |

---

## 5. Acceptance Criteria

- [x] `Generate()` returns a 6-character string containing only decimal digits `[0-9]`.
- [x] Leading zeros are preserved (e.g., `"007432"`).
- [x] 1 000 consecutive calls produce no duplicates (statistical sanity check).
- [x] `Verify(plain, hash)` returns `true` when plain is the code used to produce hash.
- [x] `Verify("wrong", hash)` returns `false`.
- [x] Test coverage for `pkg/otp` = 100%.
- [x] `golangci-lint run ./pkg/otp/...` passes with zero issues.
- [x] `gosec ./pkg/otp/...` passes with zero issues.
- [x] `docs/ROADMAP.md` row 1.1.5 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                        | Type     | Status     |
| ------------------------------------------------- | -------- | ---------- |
| `golang.org/x/crypto/bcrypt` added to `go.mod`   | External | 🔵 backlog |
| Phase 0 complete (module scaffolded)              | Upstream | ✅ done   |

---

## 7. Testing Plan

### Unit tests (`pkg/otp/otp_test.go`, no build tag)

- **Format:** assert returned `plain` matches `^[0-9]{6}$`.
- **Leading zeros:** generate 10 000 codes; assert at least one starts with `"0"` (probabilistic).
- **Hash not empty:** assert `hash` is a non-empty bcrypt hash string starting with `$2a$`.
- **Plain ≠ hash:** assert `plain != hash`.
- **Verify correct:** `Verify(plain, hash)` returns `true`.
- **Verify wrong:** `Verify("000000", hash)` returns `false` (when hash is not for `"000000"`).
- **Uniqueness:** 1 000 calls; deduplicate; assert len == 1 000.

### Integration tests

N/A

---

## 8. Open Questions

| # | Question                                           | Owner | Resolution |
| - | -------------------------------------------------- | ----- | ---------- |
| 1 | Should `bcryptCost` be configurable via `pkg/config`? | — | Export a `SetCost(n int)` or accept cost as a parameter in `Generate` — keeps the package testable without heavy bcrypt cost in test runs. |

---

## 9. Change Log

| Date       | Author | Change                        |
| ---------- | ------ | ----------------------------- |
| 2026-03-07 | —      | Task created from roadmap 1.1.5 |
