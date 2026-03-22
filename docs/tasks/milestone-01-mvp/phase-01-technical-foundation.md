# Task 1.1.0 — Technical Foundation & Engineering Standards

> **Roadmap Ref:** Phase 1 — Technical Foundation & Standards
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-22
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Establish the core architecture, development standards, and infrastructure plumbing for Moolah. This includes setting up the Go environment, observability, and database connectivity.

---

## 2. Context & Motivation

Before building financial features, we need a robust, traceable, and type-safe foundation to ensure the "Always Cents" rule and multi-entity logic are handled consistently.

- Reference: `docs/design/003-moolah-product.md` (Architecture section)
- Link to relevant roadmap row: `docs/tasks/roadmap.md#Phase-1`

---

## 3. Scope

### In scope

- [ ] Go repository initialization and layout.
- [ ] Docker-compose for PostgreSQL 17 and Redis 7.
- [ ] Structured logging (`slog`) and ULID-based Request ID middleware.
- [ ] OpenTelemetry (OTel) basic trace propagation.
- [ ] SQL migration setup (`goose`) and `sqlc` configuration.

### Out of scope

- Business domain entities (Currencies, Accounts).
- UI implementation.

---

## 4. Technical Design

### Files to create / modify

| Action   | Path                                      | Purpose                       |
| -------- | ----------------------------------------- | ----------------------------- |
| CREATE   | `cmd/api/main.go`                         | Entry point                   |
| CREATE   | `internal/platform/log/slog.go`           | Logger configuration          |
| CREATE   | `internal/api/middleware/request_id.go`    | ULID Request ID logic         |
| CREATE   | `sqlc.yaml`                               | sqlc configuration            |
| CREATE   | `docker-compose.yml`                      | Local infrastructure          |

---

## 5. Acceptance Criteria

- [ ] `docker-compose up` spins up Postgres and Redis.
- [ ] Every request log entry includes a `request_id`.
- [ ] `sqlc generate` works with a sample query.
- [ ] `go build` succeeds.
