# Moolah — Project Roadmap

> **Version:** 1.0.0 | **Last Updated:** 2026-03-12 | **Status:** 🟡 In Progress

---

## Legend

| Badge | Meaning |
| --- | --- |
| 🔵 `backlog` | Planned but not yet started |
| 🟡 `in-progress` | Actively being worked on |
| ✅ `done` | Completed and merged |
| ❌ `canceled` | Dropped — will not be implemented |
| ⏸️ `postponed` | Deferred to a future phase |
| 🚫 `blocked` | Cannot proceed; dependency or decision pending |

---

## Phase 0 — Foundation & Architecture

> **Goal:** Establish the project skeleton, core tooling, and architectural decisions before writing any business logic.
> **Status:** ✅ `done` | **Last Updated:** 2026-03-12

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 0.1 | Define system architecture & ADRs | ✅ `done` | 2026-03-06 | `docs/ARCHITECTURE.md` |
| 0.2 | Define consolidated DDL schema | ✅ `done` | 2026-03-06 | `docs/schema.sql` |
| 0.3 | Define project roadmap | ✅ `done` | 2026-03-06 | `docs/ROADMAP.md` (this file) |
| 0.4 | Initialize Go module & directory layout | ✅ `done` | 2026-03-06 | `go mod init`; scaffolded per ADR |
| 0.5 | Configure `golangci-lint` (`.golangci.yml`) | ✅ `done` | 2026-03-06 | v2 config added at repo root; excludes generated sqlc/swagger outputs |
| 0.6 | Configure `sqlc` (`sqlc.yaml`) | ✅ `done` | 2026-03-06 | Named params (`@name`); `emit_interface: true` |
| 0.7 | Configure `Makefile` (lint, test, generate, run) | ✅ `done` | 2026-03-06 | |
| 0.8 | Configure `docker-compose.yml` (Postgres + Redis) | ✅ `done` | 2026-03-06 | Local dev environment |
| 0.9 | Write `Dockerfile` (multi-stage production build) | ✅ `done` | 2026-03-06 | |
| 0.10 | Set up GitHub Actions CI pipeline (`ci.yml`) | ✅ `done` | 2026-03-06 | Lint → Security → Unit → Integration → Build → Release (release-please) |

---

## Phase 1 — MVP: Core Finance (Accounts Payable & Cash Flow)

> **Goal:** Deliver a fully functional, production-ready API covering Tenants, Users, Accounts, Categories, and Transactions. This is the revenue-enabling phase.
> **Status:** ✅ `done` | **Last Updated:** 2026-03-12

### 1.1 Infrastructure & Platform

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 1.1.1 | `pkg/ulid` — thread-safe monotonic ULID factory | ✅ `done` | 2026-03-07 | |
| 1.1.2 | `pkg/logger` — structured `slog` JSON logger | ✅ `done` | 2026-03-07 | |
| 1.1.3 | `pkg/config` — env-based config with validation | ✅ `done` | 2026-03-07 | Panic on missing required vars |
| 1.1.4 | `pkg/paseto` — PASETO v4 local seal/parse | ✅ `done` | 2026-03-07 | `aidanwoods.dev/go-paseto` |
| 1.1.5 | `pkg/otp` — cryptographically secure 6-digit OTP | ✅ `done` | 2026-03-07 | bcrypt hash storage |
| 1.1.6 | Goose migration files (enums → tenants → users → accounts → categories → transactions → audit_logs) | ✅ `done` | 2026-03-07 | `embed.FS`; auto-run on startup |
| 1.1.7 | `sqlc` query files for all Phase 1 entities | ✅ `done` | 2026-03-07 | |
| 1.1.8 | `sqlc generate` — verified generated code committed | ✅ `done` | 2026-03-07 | Checked in CI |
| 1.1.9 | `platform/middleware/auth.go` — PASETO validation + context injection | ✅ `done` | 2026-03-07 | `RequireAuth`, `RequireRole` |
| 1.1.10 | `platform/middleware/ratelimit.go` — token-bucket OTP limiter | ✅ `done` | 2026-03-07 | `golang.org/x/time/rate`; 5 req/15 min |
| 1.1.11 | `platform/middleware/logger.go` — request logging middleware | ✅ `done` | 2026-03-07 | tenant_id, user_id, latency |
| 1.1.12 | `platform/mailer/smtp_mailer.go` — implements `domain.Mailer` | ✅ `done` | 2026-03-07 | |
| 1.1.13 | `platform/mailer/smtp_mailer_integration_test.go` — Testcontainers + Mailhog | ✅ `done` | 2026-03-07 | |
| 1.1.14 | `platform/middleware/idempotency.go` — Redis-backed `Idempotency-Key` middleware | ✅ `done` | 2026-03-07 | 24 h TTL; scoped per `userID`; `IdempotencyStore` interface for mockability |
| 1.1.15 | `platform/idempotency/redis_store.go` — `IdempotencyStore` Redis implementation | ✅ `done` | 2026-03-07 | `SETNX` lock + `SET` response; requires `github.com/redis/go-redis/v9` |
| 1.1.16 | `internal/testutil/containers` — centralized testcontainers-go helpers (Postgres, Redis, Mailhog) | ✅ `done` | 2026-03-07 | Shared via `TestMain`; `//go:build integration`; eliminates per-test container setup |
| 1.1.17 | `internal/testutil/mocks` — centralized testify/mock implementations (Querier, IdempotencyStore, Mailer) | ✅ `done` | 2026-03-12 | Updates to include Master Purchase queries. |
| 1.1.18 | `internal/testutil/seeds` — canonical test-data factories (tenant, user, account, category, transaction) | ✅ `done` | 2026-03-07 | `//go:build integration`; used by repository and service integration tests |
| 1.1.19 | `internal/platform/bootstrap/sysadmin.go` — idempotent sysadmin bootstrap on startup (`SYSADMIN_EMAIL` env var) | ✅ `done` | 2026-03-10 | Breaks bootstrap paradox: creates system tenant + sysadmin user if absent; no-op on subsequent starts |

---

## Phase 2 — Credit Card & Installment Tracking

> **Goal:** Introduce credit card accounts with the "Master Purchase" installment model — one record per purchase, physical transaction rows created only at invoice-close time. Keeps the DB lean and projections at runtime.
> **Status:** 🟡 `in-progress` | **Last Updated:** 2026-03-12

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 2.1 | `domain/master_purchase.go` — entity + repository interface | ✅ `done` | 2026-03-12 | Task 2.1 |
| 2.2 | [DB Migration: `master_purchases`](tasks/TASK_2.2_db-migration-master-purchases.md) | ✅ `done` | 2026-03-12 | Task 2.2 |
| 2.3 | [SQLC Queries for Master Purchases](tasks/TASK_2.3_sqlc-queries-master-purchase.md) | ✅ `done` | 2026-03-12 | Task 2.3 |
| 2.4 | [MasterPurchaseRepository Implementation](tasks/TASK_2.4_repository-master-purchase.md) | ✅ `done` | 2026-03-12 | Task 2.4 |
| 2.5 | [MasterPurchase Service Layer](tasks/TASK_2.5_service-master-purchase.md) | ✅ `done` | 2026-03-12 | Task 2.5 |
| 2.6 | [MasterPurchase HTTP Handlers](tasks/TASK_2.6_handler-master-purchase.md) | ✅ `done` | 2026-03-12 | Task 2.6 |
| 2.7 | [InvoiceCloser Service & System Actor](tasks/TASK_2.7_service-invoice-closer.md) | ✅ `done` | 2026-03-12 | Task 2.7 |
| 2.8 | Endpoint `POST /v1/accounts/:id/close-invoice` (manual trigger) | 🔵 `backlog` | 2026-03-12 | |
| 2.9 | Integration tests for invoice closing flow | 🔵 `backlog` | 2026-03-12 | |
| 2.10 | Phase 2 API smoke test — `internal/server/smoke_test.go` | 🔵 `backlog` | 2026-03-12 | |

---

## Phase 3 — Investment Portfolio Tracking

> **Goal:** Add investment accounts with position tracking, asset allocation views, and monthly performance snapshots.
> **Status:** 🔵 `backlog` | **Last Updated:** 2026-03-06

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 3.1 | ADR: investment data model | 🔵 `backlog` | 2026-03-06 | |
| 3.2 | Goose migrations — `assets`, `positions` | 🔵 `backlog` | 2026-03-06 | |

---

## Phase 4 — Billing, Plans & Monetisation

> **Goal:** Subscriptions and plan enforcement.
> **Status:** 🔵 `backlog` | **Last Updated:** 2026-03-06

---

## Phase 5 — Observability & Production Hardening
