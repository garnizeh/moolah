# Moolah — Project Roadmap

> **Version:** 1.0.0 | **Last Updated:** 2026-03-07 | **Status:** 🟡 In Progress
> **Version:** 1.0.0 | **Last Updated:** 2026-03-08 | **Status:** 🟡 In Progress
| 1.5.2 | `cmd/api/server.go` — `http.Server` factory, middleware chain | ✅ `done` | 2026-03-08 | Implemented in `internal/server`: global logger middleware applied; `/healthz` route fixed. |

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
> **Status:** 🟡 `in-progress` | **Last Updated:** 2026-03-06

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
> **Status:** � `in-progress` | **Last Updated:** 2026-03-07

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
| 1.1.17 | `internal/testutil/mocks` — centralized testify/mock implementations (Querier, IdempotencyStore, Mailer) | ✅ `done` | 2026-03-07 | Centralized mocks implemented; tests updated to use `internal/testutil/mocks`. |
| 1.1.18 | `internal/testutil/seeds` — canonical test-data factories (tenant, user, account, category, transaction) | ✅ `done` | 2026-03-07 | `//go:build integration`; used by repository and service integration tests |

### 1.2 Domain Layer

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 1.2.1 | `domain/errors.go` — sentinel errors (`ErrNotFound`, `ErrForbidden`, `ErrInvalidOTP`, …) | ✅ `done` | 2026-03-07 | |
| 1.2.2 | `domain/role.go` — `Role` type + constants | ✅ `done` | 2026-03-07 | |
| 1.2.3 | `domain/tenant.go` — entity + `TenantRepository` interface | ✅ `done` | 2026-03-07 | |
| 1.2.4 | `domain/user.go` — entity + `UserRepository` interface | ✅ `done` | 2026-03-07 | |
| 1.2.5 | `domain/auth.go` — `OTPRequest` entity + `AuthRepository` interface | ✅ `done` | 2026-03-07 | |
| 1.2.6 | `domain/mailer.go` — `Mailer` interface | ✅ `done` | 2026-03-07 | |
| 1.2.7 | `domain/account.go` — entity + `AccountRepository` interface | ✅ `done` | 2026-03-07 | |
| 1.2.8 | `domain/category.go` — entity + `CategoryRepository` interface | ✅ `done` | 2026-03-07 | |
| 1.2.9 | `domain/transaction.go` — entity + `TransactionRepository` interface | ✅ `done` | 2026-03-07 | |
| 1.2.10 | `domain/audit.go` — `AuditLog` entity + `AuditRepository` interface | ✅ `done` | 2026-03-07 | |
| 1.2.11 | `domain/admin.go` — admin-only repository interfaces | ✅ `done` | 2026-03-07 | |

### 1.3 Repository Layer

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 1.3.1 | `repository/tenant_repo.go` | ✅ `done` | 2026-03-07 | 100% coverage, shared mock |
| 1.3.2 | `repository/user_repo.go` | ✅ `done` | 2026-03-07 | 100% coverage, auth-flow exception |
| 1.3.3 | `repository/auth_repo.go` | ✅ `done` | 2026-03-07 | 100% coverage, OTP lifecycle |
| 1.3.4 | `repository/account_repo.go` | ✅ `done` | 2026-03-07 | 100% coverage |
| 1.3.5 | `repository/category_repo.go` | ✅ `done` | 2026-03-07 | 100% coverage |
| 1.3.6 | `repository/transaction_repo.go` | ✅ `done` | 2026-03-07 | 100% coverage |
| 1.3.7 | `repository/audit_repo.go` | ✅ `done` | 2026-03-07 | |
| 1.3.8 | `repository/admin_repo.go` | ✅ `done` | 2026-03-07 | |
| 1.3.9 | Integration tests for all repositories (testcontainers-go) | ✅ `done` |

### 1.4 Service Layer

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 1.4.1 | `service/auth_service.go` + unit tests | ✅ `done` | 2026-03-07 | ReqOTP, VerifyOTP, RefreshToken. 81% cov. |
| 1.4.2 | `service/tenant_service.go` + unit tests | ✅ `done` | 2026-03-07 | CRUD, invite user |
| 1.4.3 | `service/account_service.go` + unit tests | ✅ `done` | 2026-03-07 | CRUD, balance recalculation |
| 1.4.4 | `service/category_service.go` + unit tests | ✅ `done` | 2026-03-07 | CRUD, hierarchy validation |
| 1.4.5 | `service/transaction_service.go` + unit tests | ✅ `done` | 2026-03-07 | CRUD, audit trail |
| 1.4.6 | `service/admin_service.go` + unit tests | ✅ `done` | 2026-03-07 | Cross-tenant ops |

### 1.5 HTTP Handler Layer

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 1.5.1 | `cmd/api/main.go` — DI wiring, Goose migrations, server start | ✅ `done` | 2026-03-07 | |
| 1.5.2 | `cmd/api/server.go` — `http.Server` factory, middleware chain | 🟡 `in-progress` | 2026-03-07 | |
| 1.5.3 | `cmd/api/routes.go` — all route registrations | 🟡 `in-progress` | 2026-03-07 | Go 1.22 `METHOD /path/{param}` syntax |
| 1.5.4 | `handler/auth_handler.go` — `RequestOTP`, `VerifyOTP`, `RefreshToken` | 🔵 `backlog` | 2026-03-07 | |
| 1.5.5 | `handler/tenant_handler.go` — `GetMe`, `UpdateMe`, `InviteUser` | 🔵 `backlog` | 2026-03-07 | |
| 1.5.6 | `handler/account_handler.go` — full CRUD | 🔵 `backlog` | 2026-03-07 | |
| 1.5.7 | `handler/category_handler.go` — CRUD | 🔵 `backlog` | 2026-03-07 | |
| 1.5.8 | `handler/transaction_handler.go` — full CRUD + list with filters | 🔵 `backlog` | 2026-03-07 | |
| 1.5.9 | `handler/admin_handler.go` — sysadmin routes | 🔵 `backlog` | 2026-03-07 | |
| 1.5.10 | Swaggo annotations on all handlers; `swag init` verified in CI | 🔵 `backlog` | 2026-03-07 | `docs/swagger/` |
| 1.5.11 | Wire `Idempotency` middleware on all mutating `POST` routes | 🔵 `backlog` | 2026-03-07 | Depends on 1.1.14 + 1.1.15; apply after `RequireAuth` in the chain |

### 1.6 Quality Gate

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 1.6.1 | ≥ 80% unit test coverage enforced in CI | 🔵 `backlog` | 2026-03-06 | |
| 1.6.2 | `govulncheck` passing in CI | 🔵 `backlog` | 2026-03-06 | |
| 1.6.3 | `gosec` passing in CI | 🔵 `backlog` | 2026-03-06 | |
| 1.6.4 | Full Phase 1 API smoke test (Postman / httpie collection) | 🔵 `backlog` | 2026-03-06 | |
| 1.6.5 | Generate Swagger documentation via Swaggo | 🔵 `backlog` | 2026-03-08 | Includes Makefile rule and CI check |
| 1.6.6 | Generate Bruno collection for API calls | 🔵 `backlog` | 2026-03-08 | Replacement for Postman in Phase 1 |

---

## Phase 2 — Credit Card & Installment Tracking

> **Goal:** Introduce credit card accounts with the "Master Purchase" installment model — one record per purchase, physical transaction rows created only at invoice-close time. Keeps the DB lean and projections at runtime.
> **Status:** 🔵 `backlog` | **Last Updated:** 2026-03-06

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 2.1 | `domain/master_purchase.go` — entity + `MasterPurchaseRepository` interface | 🔵 `backlog` | 2026-03-06 | |
| 2.2 | Goose migration — `master_purchases` table | 🔵 `backlog` | 2026-03-06 | |
| 2.3 | `sqlc` queries for `master_purchases` | 🔵 `backlog` | 2026-03-06 | |
| 2.4 | `repository/master_purchase_repo.go` + integration tests | 🔵 `backlog` | 2026-03-06 | |
| 2.5 | `service/master_purchase_service.go` + unit tests | 🔵 `backlog` | 2026-03-06 | Create, project installments at runtime |
| 2.6 | `handler/master_purchase_handler.go` — `POST /v1/master-purchases` | 🔵 `backlog` | 2026-03-06 | Returns projected installments in response |
| 2.7 | `service/invoice_closer.go` — scheduled job / on-demand trigger | 🔵 `backlog` | 2026-03-06 | Materialises installment `transactions` at closing date |
| 2.8 | Endpoint `POST /v1/accounts/:id/close-invoice` (manual trigger) | 🔵 `backlog` | 2026-03-06 | |
| 2.9 | Remainder-cent handling in installment calculation | 🔵 `backlog` | 2026-03-06 | Last installment absorbs rounding remainder |
| 2.10 | Audit trail for `SYSTEM` actor on auto-generated transactions | 🔵 `backlog` | 2026-03-06 | `actor_id = "SYSTEM"` |
| 2.11 | Update Swagger docs; `swag init` in CI | 🔵 `backlog` | 2026-03-06 | |
| 2.12 | Integration tests for invoice closing flow | 🔵 `backlog` | 2026-03-06 | |

---

## Phase 3 — Investment Portfolio Tracking

> **Goal:** Add investment accounts with position tracking, asset allocation views, and monthly performance snapshots. Read-heavy; no monetary mutation beyond deposits/withdrawals recorded as transactions.
> **Status:** 🔵 `backlog` | **Last Updated:** 2026-03-06

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 3.1 | ADR: investment data model (positions, assets, snapshots) | 🔵 `backlog` | 2026-03-06 | |
| 3.2 | Goose migrations — `assets`, `positions`, `portfolio_snapshots` | 🔵 `backlog` | 2026-03-06 | |
| 3.3 | `domain/investment.go` — entities + repository interfaces | 🔵 `backlog` | 2026-03-06 | |
| 3.4 | `sqlc` query files for investment entities | 🔵 `backlog` | 2026-03-06 | |
| 3.5 | `repository/investment_repo.go` + integration tests | 🔵 `backlog` | 2026-03-06 | |
| 3.6 | `service/investment_service.go` + unit tests | 🔵 `backlog` | 2026-03-06 | Position upsert, allocation calc |
| 3.7 | `handler/investment_handler.go` — positions, allocation, history | 🔵 `backlog` | 2026-03-06 | |
| 3.8 | Monthly snapshot job (`portfolio_snapshots`) | 🔵 `backlog` | 2026-03-06 | Triggered by scheduler or cron |
| 3.9 | `GET /v1/investments/summary` — net worth + allocation breakdown | 🔵 `backlog` | 2026-03-06 | |
| 3.10 | Currency conversion hook (extensible; no external API in MVP) | 🔵 `backlog` | 2026-03-06 | Store rates manually or via future integration |

---

## Phase 4 — Billing, Plans & Monetisation

> **Goal:** Implement subscription plan enforcement (`free`/`basic`/`premium` tiers), usage quotas, and integration with a payment gateway (Stripe or equivalent).
> **Status:** 🔵 `backlog` | **Last Updated:** 2026-03-06

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 4.1 | ADR: billing strategy (Stripe / Paddle / manual invoice) | 🔵 `backlog` | 2026-03-06 | |
| 4.2 | Plan quota enforcement middleware (account/transaction limits) | 🔵 `backlog` | 2026-03-06 | |
| 4.3 | Webhook handler for payment gateway events | 🔵 `backlog` | 2026-03-06 | |
| 4.4 | `POST /v1/tenants/me/subscription` — upgrade/downgrade | 🔵 `backlog` | 2026-03-06 | |
| 4.5 | Grace-period logic on plan downgrade | 🔵 `backlog` | 2026-03-06 | |
| 4.6 | Admin dashboard endpoint: MRR, churn, plan distribution | 🔵 `backlog` | 2026-03-06 | |

---

## Phase 5 — Observability & Production Hardening

> **Goal:** Ship the monitoring, alerting, and reliability features required to operate at scale with confidence.
> **Status:** 🔵 `backlog` | **Last Updated:** 2026-03-06

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 5.1 | `/healthz` and `/readyz` endpoints | 🔵 `backlog` | 2026-03-06 | Readiness checks DB + Redis |
| 5.2 | Prometheus metrics middleware (`/metrics` scrape endpoint) | 🔵 `backlog` | 2026-03-06 | Request count, latency histograms |
| 5.3 | OpenTelemetry tracing (OTLP exporter) | 🔵 `backlog` | 2026-03-06 | |
| 5.4 | Centralized error tracking (Sentry or equivalent) | 🔵 `backlog` | 2026-03-06 | |
| 5.5 | PgBouncer connection pooling config | 🔵 `backlog` | 2026-03-06 | Phase 2 scaling path from ARCHITECTURE.md |
| 5.6 | Redis Sentinel config for HA | 🔵 `backlog` | 2026-03-06 | |
| 5.7 | PostgreSQL read replicas for report queries | 🔵 `backlog` | 2026-03-06 | |
| 5.8 | Runbook + on-call playbook in `docs/` | 🔵 `backlog` | 2026-03-06 | |

---

## Decisions & Deferred Items

| Item | Decision | Rationale | Status |
| --- | --- | --- | --- |
| External router (chi, gorilla/mux) | ❌ **Rejected** | Go 1.22 stdlib routing covers all needs; zero-dependency preferred | ❌ `canceled` |
| GraphQL API | ⏸️ **Deferred** | REST covers Phase 1–3 requirements; revisit if client demand justifies complexity | ⏸️ `postponed` |
| gRPC internal transport | ⏸️ **Deferred** | Monolith for now; re-evaluate at Phase 5 if microservices split occurs | ⏸️ `postponed` |
| Real-time push (WebSocket / SSE) | ⏸️ **Deferred** | Phase 3+ feature for live portfolio updates | ⏸️ `postponed` |
| Mobile SDK / OpenAPI client gen | ⏸️ **Deferred** | Swagger spec is generated; client gen tooling deferred | ⏸️ `postponed` |
| GORM / heavy ORM | ❌ **Rejected** | `sqlc` + raw SQL is explicit, auditable, and type-safe | ❌ `canceled` |
| Float for monetary values | ❌ **Rejected** | `int64` cents only; floating-point drift is unacceptable for finance | ❌ `canceled` |

---

> ⚠️ **Maintenance Contract:** This document **must** be updated whenever a task or phase changes state. Every row must carry an accurate `Last Updated` date. Stale roadmap entries are treated as bugs.
