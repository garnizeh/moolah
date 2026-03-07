# Task 1.1.2 — pkg/logger: Structured slog JSON Logger

> **Roadmap Ref:** Phase 1 — MVP: Core Finance › 1.1 Infrastructure & Platform
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement a thin package that constructs a `*slog.Logger` configured for structured JSON output (production) or human-readable text output (development), driven by the environment. All other packages receive this logger via dependency injection; nothing calls `slog.SetDefault`.

---

## 2. Context & Motivation

Go 1.21+ ships `log/slog` in the standard library. Using it directly gives us structured, levelled, JSON-capable logging with zero external dependencies. The package must be initialised once at startup and passed down the dependency graph — never used as a global.

- Roadmap row: `docs/ROADMAP.md` › Phase 1 › 1.1.2
- Depends on: task 1.1.3 (`pkg/config`) for the log level and format settings, but can be initialised independently with explicit parameters.

---

## 3. Scope

### In scope

- [ ] `pkg/logger/logger.go` — `New(level, format string) *slog.Logger`
- [ ] `format` values: `"json"` (default/prod) and `"text"` (local dev)
- [ ] `level` values: `"debug"`, `"info"`, `"warn"`, `"error"` (default `"info"`)
- [ ] `pkg/logger/logger_test.go` — unit tests verifying level filtering and output format

### Out of scope

- Log sampling / rate-limiting (deferred to Phase 5 observability)
- OpenTelemetry log bridge (deferred to task 5.3)
- Rotating file sinks (stdout/stderr only in Phase 1)

---

## 4. Technical Design

### Files to create / modify

| Action | Path                        | Purpose                                    |
| ------ | --------------------------- | ------------------------------------------ |
| CREATE | `pkg/logger/logger.go`      | Logger factory using stdlib `log/slog`     |
| CREATE | `pkg/logger/logger_test.go` | Unit tests for level filtering and formats |

### Key interfaces / types

```go
// pkg/logger/logger.go
package logger

import (
    "log/slog"
    "os"
    "strings"
)

// New returns a *slog.Logger configured for the given level and format.
// level:  "debug" | "info" | "warn" | "error"  (default: "info")
// format: "json"  | "text"                     (default: "json")
func New(level, format string) *slog.Logger {
    var lvl slog.Level
    switch strings.ToLower(level) {
    case "debug":
        lvl = slog.LevelDebug
    case "warn":
        lvl = slog.LevelWarn
    case "error":
        lvl = slog.LevelError
    default:
        lvl = slog.LevelInfo
    }

    opts := &slog.HandlerOptions{Level: lvl}

    var handler slog.Handler
    if strings.ToLower(format) == "text" {
        handler = slog.NewTextHandler(os.Stdout, opts)
    } else {
        handler = slog.NewJSONHandler(os.Stdout, opts)
    }

    return slog.New(handler)
}
```

### SQL queries (sqlc)

N/A

### API endpoints (if applicable)

N/A

### Error cases to handle

| Scenario              | Handling                                          |
| --------------------- | ------------------------------------------------- |
| Unknown `level` value | Fall back to `slog.LevelInfo`; log a warning once |
| Unknown `format` value| Fall back to JSON handler                         |

---

## 5. Acceptance Criteria

- [ ] `New("info", "json")` returns a `*slog.Logger` whose output is valid JSON.
- [ ] `New("info", "text")` returns a `*slog.Logger` whose output is human-readable text.
- [ ] Messages below the configured level are suppressed.
- [ ] The function never calls `slog.SetDefault` or writes to a global.
- [ ] Test coverage for `pkg/logger` = 100%.
- [ ] `golangci-lint run ./pkg/logger/...` passes with zero issues.
- [ ] `gosec ./pkg/logger/...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row 1.1.2 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                          | Type     | Status     |
| ----------------------------------- | -------- | ---------- |
| Go 1.21+ (`log/slog` stdlib)        | Runtime  | ✅ done   |
| Phase 0 complete (module scaffolded)| Upstream | ✅ done   |

---

## 7. Testing Plan

### Unit tests (`pkg/logger/logger_test.go`, no build tag)

- Write output to a `bytes.Buffer` via a custom handler, not stdout, to allow assertions.
- **JSON format:** parse buffer output as JSON; assert `level` and `msg` keys are present.
- **Text format:** assert output contains the level string and message.
- **Level filtering — debug suppressed at info level:** emit a Debug log; assert buffer is empty.
- **Level filtering — error emitted at info level:** emit an Error log; assert buffer has content.
- **Fallback level:** call `New("invalid", "json")`; assert returned logger uses `LevelInfo`.

### Integration tests

N/A

---

## 8. Open Questions

| # | Question                                          | Owner | Resolution |
| - | ------------------------------------------------- | ----- | ---------- |
| 1 | Should `New` accept `io.Writer` for testability? | —     | Yes — add an optional writer parameter to avoid capturing stdout in tests. |

---

## 9. Change Log

| Date       | Author | Change                        |
| ---------- | ------ | ----------------------------- |
| 2026-03-07 | —      | Task created from roadmap 1.1.2 |
