# Task 1.5.1 — `cmd/api/main.go` — DI wiring, Goose migrations, server start

> **Roadmap Ref:** Phase 1 — MVP › 1.5 HTTP Handler Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the application entry point at `cmd/api/main.go`. This file is the Dependency Injection root: it loads configuration, connects to Postgres and Redis, runs Goose migrations, wires all repositories and services, and starts the HTTP server. It is the only file in the project that knows about every concrete dependency.

---

## 2. Context & Motivation

The project has a fully implemented domain, repository, and service layer, but no runnable binary. This task creates the executable that ties everything together. Following the pragmatic DDD directory layout defined in `docs/ARCHITECTURE.md`, `cmd/api/` is the entry point and the only layer allowed to perform direct dependency injection. See roadmap item 1.5.1.

---

## 3. Scope

### In scope

- [x] `cmd/api/main.go` — entry point: config load, DB connect, Redis connect, migration run, DI, server start.
- [x] Graceful shutdown: listen for `SIGINT`/`SIGTERM`; drain in-flight requests before exit.
- [x] Run Goose migrations on startup via the embedded migration FS.
- [x] Wire all repositories from `internal/platform/repository/`.
- [x] Wire all services from `internal/service/`.
- [ ] Wire the server defined in Task 1.5.2 (`cmd/api/server.go`).

### Out of scope

- `server.go` and `routes.go` — Tasks 1.5.2 and 1.5.3.
- Handler implementations — Tasks 1.5.4 through 1.5.9.
- Health/readiness endpoints — Task 5.1.

---

## 4. Technical Design

### Files to create / modify

| Action | Path               | Purpose                                                    |
| ------ | ------------------ | ---------------------------------------------------------- |
| CREATE | `cmd/api/main.go`  | Application entry point, DI root, graceful shutdown        |

### Key design decisions

- Use `pkg/config` to load all environment variables; panic on missing required vars.
- Use `jackc/pgx/v5/pgxpool` for Postgres connection pooling.
- Use `redis/go-redis/v9` for Redis connection.
- Pass `*slog.Logger` as the single structured logger throughout the DI graph.
- Graceful shutdown timeout: 30 seconds.
- All error handling in `main` must be logged and cause `os.Exit(1)` — no silent failures.

### Wiring order

1. Load config (`pkg/config`).
2. Init logger (`pkg/logger`).
3. Connect DB (`pgxpool.New`).
4. Run migrations (`goose.Up`).
5. Connect Redis for idempotency store.
6. Construct repositories.
7. Construct services.
8. Construct handlers.
9. Build server (Task 1.5.2) with routes (Task 1.5.3).
10. Start server; block on `os.Signal`.
11. On signal: call `server.Shutdown(ctx)` with 30s timeout.

---

## 5. Acceptance Criteria

- [ ] `go build ./cmd/api/` compiles cleanly.
- [ ] Application starts, runs migrations, and listens on the configured port.
- [ ] Graceful shutdown drains requests within 30 seconds.
- [ ] Missing required env vars cause exit with a clear error message.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                    | Type     | Status  |
| --------------------------------------------- | -------- | ------- |
| Task 1.5.2 — `cmd/api/server.go`              | Upstream | 🔵 backlog |
| Task 1.5.3 — `cmd/api/routes.go`              | Upstream | 🔵 backlog |
| All 1.4.x Service tasks                        | Upstream | ✅ done |
| All 1.3.x Repository tasks                     | Upstream | ✅ done |
| Task 1.1.3 — `pkg/config`                      | Upstream | ✅ done |
| Task 1.1.2 — `pkg/logger`                      | Upstream | ✅ done |
| Task 1.1.6 — Goose migrations                  | Upstream | ✅ done |
| Task 1.1.14 — Idempotency middleware            | Upstream | ✅ done |
| Task 1.1.15 — Redis idempotency store           | Upstream | ✅ done |

---

## 7. Testing Plan

### Unit tests

N/A — `main.go` is the DI root and is not unit-testable in isolation.

### Integration / smoke test

- `make run` starts the server without errors.
- `GET /healthz` (Task 5.1) returns `200 OK` after wiring.
- Smoke tests in Task 1.6.4 exercise the full wired stack.

---

## 8. Open Questions

| # | Question                                              | Owner | Resolution |
| - | ----------------------------------------------------- | ----- | ---------- |
| 1 | Should connection string be a DSN or individual env vars? | — | Use single `DATABASE_URL` DSN for simplicity. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
