# Task 1.5.2 — `cmd/api/server.go` — `http.Server` factory, middleware chain

> **Roadmap Ref:** Phase 1 — MVP › 1.5 HTTP Handler Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement `cmd/api/server.go`, which constructs the `*http.Server` with timeouts, applies the global middleware chain, and returns a ready-to-start server. It is a factory function called from `main.go` after all dependencies are wired.

---

## 2. Context & Motivation

Separating server construction from `main.go` improves testability and readability. All timeout configuration, middleware ordering, and `http.ServeMux` assembly live here. See `docs/ARCHITECTURE.md` and roadmap item 1.5.2.

---

## 3. Scope

### In scope

- [ ] `cmd/api/server.go` — `NewServer(cfg *config.Config, handler http.Handler, logger *slog.Logger) *http.Server`.
- [ ] Configure `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `ReadHeaderTimeout`.
- [ ] Define the global middleware chain (applied in order):
  1. Logger middleware (`platform/middleware/logger.go`)
  2. Auth middleware guard functions available for per-route use (not global).
- [ ] No business logic; purely infrastructure assembly.

### Out of scope

- Route registration — Task 1.5.3.
- Handler implementations — Tasks 1.5.4–1.5.9.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                 | Purpose                                          |
| ------ | -------------------- | ------------------------------------------------ |
| CREATE | `cmd/api/server.go`  | `http.Server` factory with timeouts and middleware |

### Server configuration

| Setting             | Value       | Source              |
| ------------------- | ----------- | ------------------- |
| `ReadTimeout`       | 10s         | hardcoded constant  |
| `WriteTimeout`      | 30s         | hardcoded constant  |
| `IdleTimeout`       | 120s        | hardcoded constant  |
| `ReadHeaderTimeout` | 5s          | hardcoded constant  |
| `Addr`              | `:PORT`     | `cfg.ServerPort`    |

### Middleware chain

```
Request → Logger → [per-route: RateLimit, Auth, Idempotency] → Handler
```

---

## 5. Acceptance Criteria

- [ ] `NewServer` returns a properly configured `*http.Server`.
- [ ] All four timeouts are set.
- [ ] Logger middleware is applied globally.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                               | Type     | Status     |
| ---------------------------------------- | -------- | ---------- |
| Task 1.5.3 — `cmd/api/routes.go`         | Downstream | 🔵 backlog |
| Task 1.1.11 — Logger middleware           | Upstream | ✅ done    |
| Task 1.1.3 — `pkg/config`                | Upstream | ✅ done    |

---

## 7. Testing Plan

### Unit tests

- Verify `NewServer` returns a non-nil `*http.Server`.
- Verify all timeouts are configured.

---

## 8. Open Questions

| # | Question                                       | Owner | Resolution |
| - | ---------------------------------------------- | ----- | ---------- |
| 1 | Apply rate-limit globally or per-route only?   | —     | Per-route only (`/auth/*` endpoints). |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
