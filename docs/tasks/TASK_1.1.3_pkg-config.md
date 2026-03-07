# Task 1.1.3 — pkg/config: Environment-Based Configuration

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement a `pkg/config` package that reads all application settings from environment variables, validates that required variables are present, and panics at startup if any required value is missing. The resulting `Config` struct is constructed once in `main` and passed down via dependency injection — no global variable is exported.

---

## 2. Context & Motivation

Twelve-factor app discipline requires all configuration to come from the environment. Panicking on missing required vars is intentional: a misconfigured instance should fail immediately at startup rather than silently misbehaving at runtime. All other Phase 1 packages (`logger`, `pasetoutils`, `mailer`, DB pool) depend on values from this config.

- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.3

---

## 3. Scope

### In scope

- [ ] `pkg/config/config.go` — `Load() *Config` (panics on missing required vars)
- [ ] `Config` struct covering: server, database, Redis, PASETO, SMTP, logging settings
- [ ] `pkg/config/config_test.go` — table-driven tests using `t.Setenv`

### Out of scope

- Hot-reload of config at runtime (not needed for Phase 1)
- Secrets manager integration (AWS SSM, Vault) — deferred to Phase 4/5
- `.env` file loading (developers use `direnv` or `make run` with exported vars)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                        | Purpose                                         |
| ------ | --------------------------- | ----------------------------------------------- |
| CREATE | `pkg/config/config.go`      | Config struct + `Load()` factory                |
| CREATE | `pkg/config/config_test.go` | Table-driven tests with `t.Setenv`              |

### Key interfaces / types

```go
// pkg/config/config.go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

type Config struct {
    // Server
    HTTPPort        string
    ReadTimeout     time.Duration
    WriteTimeout    time.Duration
    ShutdownTimeout time.Duration

    // Database
    DatabaseURL     string

    // Redis
    RedisAddr       string
    RedisPassword   string

    // PASETO
    PasetoSecretKey string // 32-byte hex-encoded symmetric key
    TokenTTL        time.Duration

    // SMTP / Mailer
    SMTPHost        string
    SMTPPort        int
    SMTPUser        string
    SMTPPassword    string
    EmailFrom       string

    // Logging
    LogLevel        string // debug | info | warn | error
    LogFormat       string // json | text
}

// Load reads environment variables and returns a populated Config.
// It panics if any required variable is missing or malformed.
func Load() *Config { ... }
```

**Required variables (panic if absent):**
`DATABASE_URL`, `REDIS_ADDR`, `PASETO_SECRET_KEY`, `SMTP_HOST`, `SMTP_USER`, `SMTP_PASSWORD`, `EMAIL_FROM`

**Optional variables (with defaults):**

| Variable            | Default   |
| ------------------- | --------- |
| `HTTP_PORT`         | `8080`    |
| `READ_TIMEOUT`      | `10s`     |
| `WRITE_TIMEOUT`     | `30s`     |
| `SHUTDOWN_TIMEOUT`  | `15s`     |
| `REDIS_PASSWORD`    | `""`      |
| `TOKEN_TTL`         | `24h`     |
| `SMTP_PORT`         | `587`     |
| `LOG_LEVEL`         | `info`    |
| `LOG_FORMAT`        | `json`    |

### SQL queries (sqlc)

N/A

### API endpoints (if applicable)

N/A

### Error cases to handle

| Scenario                            | Handling                                     |
| ----------------------------------- | -------------------------------------------- |
| Required env var missing            | `panic(fmt.Sprintf("config: %s is required", key))` |
| Duration env var unparseable        | `panic` with descriptive message             |
| `SMTP_PORT` not a valid integer     | `panic` with descriptive message             |

---

## 5. Acceptance Criteria

- [ ] `Load()` returns a fully populated `*Config` when all required env vars are set.
- [ ] `Load()` panics with a clear message naming the missing variable when any required var is absent.
- [ ] All optional vars use the documented defaults when not set.
- [ ] No global `Config` variable is exported from the package.
- [ ] Test coverage for `pkg/config` ≥ 90% (panic paths covered via `recover`).
- [ ] `golangci-lint run ./pkg/config/...` passes with zero issues.
- [ ] `gosec ./pkg/config/...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 1.1.3 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                           | Type     | Status     |
| ------------------------------------ | -------- | ---------- |
| Go stdlib only (`os`, `strconv`, `time`) | Runtime | ✅ done |
| Phase 0 complete (module scaffolded) | Upstream | ✅ done   |

---

## 7. Testing Plan

### Unit tests (`pkg/config/config_test.go`, no build tag)

Use `t.Setenv` to isolate each test (automatically undone after test).

- **All required vars set:** `Load()` returns valid `*Config` with correct values.
- **Each required var missing individually:** wrap `Load()` in a goroutine with `recover`; assert panic message contains the var name.
- **Optional var absent:** assert default values are applied.
- **Invalid duration string:** assert panic.
- **Invalid `SMTP_PORT`:** assert panic.

### Integration tests

N/A

---

## 8. Open Questions

| # | Question                                                   | Owner | Resolution |
| - | ---------------------------------------------------------- | ----- | ---------- |
| 1 | Should we use `go-playground/validator` struct tags here? | —     | No — keep config loading dependency-free; validation is done inline in `Load()`. |

---

## 9. Change Log

| Date       | Author | Change                        |
| ---------- | ------ | ----------------------------- |
| 2026-03-07 | —      | Task created from roadmap 1.1.3 |
