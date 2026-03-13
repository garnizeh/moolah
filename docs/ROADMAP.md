# Moolah — Project Roadmap

> **Version:** 1.0.0 | **Last Updated:** 2026-03-13 | **Status:** 🟡 In Progress

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
> **Status:** ✅ `done` | **Last Updated:** 2026-03-13

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 2.1 | `domain/master_purchase.go` — entity + repository interface | ✅ `done` | 2026-03-12 | Task 2.1 |
| 2.2 | [DB Migration: `master_purchases`](tasks/TASK_2.2_db-migration-master-purchases.md) | ✅ `done` | 2026-03-12 | Task 2.2 |
| 2.3 | [SQLC Queries for Master Purchases](tasks/TASK_2.3_sqlc-queries-master-purchase.md) | ✅ `done` | 2026-03-12 | Task 2.3 |
| 2.4 | [MasterPurchaseRepository Implementation](tasks/TASK_2.4_repository-master-purchase.md) | ✅ `done` | 2026-03-12 | Task 2.4 |
| 2.5 | [MasterPurchase Service Layer](tasks/TASK_2.5_service-master-purchase.md) | ✅ `done` | 2026-03-12 | Task 2.5 |
| 2.6 | [MasterPurchase HTTP Handlers](tasks/TASK_2.6_handler-master-purchase.md) | ✅ `done` | 2026-03-12 | Task 2.6 |
| 2.7 | [InvoiceCloser Service & System Actor](tasks/TASK_2.7_service-invoice-closer.md) | ✅ `done` | 2026-03-12 | Task 2.7 |
| 2.8 | Endpoint `POST /v1/accounts/:id/close-invoice` (manual trigger) | ✅ `done` | 2026-03-12 | |
| 2.9 | [Remainder-Cent Handling](tasks/TASK_2.9_installment-remainder-cent.md) | ✅ `done` | 2026-03-12 | |
| 2.10 | [Audit System Actor Integration](tasks/TASK_2.10_audit-system-actor.md) | ✅ `done` | 2026-03-12 | |
| 2.11 | [Swagger Documentation Update](tasks/TASK_2.11_swagger-update.md) | ✅ `done` | 2026-03-13 | |
| 2.12 | [Integration tests for invoice closing flow](tasks/done/TASK_2.12_integration-tests-invoice.md) | ✅ `done` | 2026-03-13 | |
| 2.13 | [Smoke Tests Phase 2](tasks/TASK_2.13_smoke-test-phase2.md) | ✅ `done` | 2026-03-13 | |

---

## Phase 3 — Investment Portfolio Tracking

> **Goal:** Add investment accounts with position tracking, asset allocation views, income receivables, and periodic portfolio snapshots. Read-heavy; no monetary mutation beyond deposits/withdrawals recorded as transactions.
> **Status:** 🟡 `in-progress` | **Last Updated:** 2026-03-13

| # | Task | Status | Last Updated | Notes |
| --- | --- | --- | --- | --- |
| 3.1  | [ADR: investment data model](tasks/TASK_3.1_adr-investment-data-model.md) | ✅ `done` | 2026-03-13 | `docs/ADR-003-investment-data-model.md` v3 |
| 3.2  | [DB migrations — all 6 tables + 3 enums](tasks/TASK_3.2_db-migrations-investment.md) | 🔵 `backlog` | 2026-03-13 | `00010`–`00014` migration files |
| 3.3  | [Domain: `Asset` + `TenantAssetConfig` entities + interfaces](tasks/TASK_3.3_domain-investment.md) | 🔵 `backlog` | 2026-03-13 | `internal/domain/asset.go` |
| 3.4  | [SQLC queries: `assets` + `tenant_asset_configs`](tasks/TASK_3.4_sqlc-queries-investment.md) | 🔵 `backlog` | 2026-03-13 | Includes COALESCE merge query (ADR §2.7) |
| 3.5  | [Repository: `AssetRepository` + `TenantAssetConfigRepository`](tasks/TASK_3.5_repository-investment.md) | 🔵 `backlog` | 2026-03-13 | |
| 3.6  | [Service: `InvestmentService` — position CRUD, allocation, receivable lifecycle](tasks/TASK_3.6_service-investment.md) | 🔵 `backlog` | 2026-03-13 | |
| 3.7  | [HTTP handlers: asset catalogue + tenant asset configs](tasks/TASK_3.7_handler-investment.md) | 🔵 `backlog` | 2026-03-13 | `GET/POST /v1/assets`, `PUT /v1/me/asset-configs/{asset_id}` |
| 3.8  | [Portfolio snapshot job (`SNAPSHOT_CRON_SCHEDULE`)](tasks/TASK_3.8_snapshot-job.md) | 🔵 `backlog` | 2026-03-13 | Default `"0 5 1 * *"` |
| 3.9  | [Domain: position family (`Position`, `PositionSnapshot`, `PositionIncomeEvent`, `PortfolioSnapshot`)](tasks/TASK_3.9_summary-endpoint.md) | 🔵 `backlog` | 2026-03-13 | `internal/domain/position.go` |
| 3.10 | [CurrencyConverter interface + static rate implementation](tasks/TASK_3.10_currency-conversion-hook.md) | 🔵 `backlog` | 2026-03-13 | Integer-cents arithmetic; no external API in MVP |
| 3.11 | [SQLC queries: position family (positions, snapshots, income events, portfolio snapshots)](tasks/TASK_3.11_sqlc-queries-position-family.md) | 🔵 `backlog` | 2026-03-13 | Includes `ListPositionsDueIncome` for income scheduler |
| 3.12 | [Repository: position family (4 repos + integration tests)](tasks/TASK_3.12_repository-position-family.md) | 🔵 `backlog` | 2026-03-13 | |
| 3.13 | [Income scheduler service (background goroutine — ADR §9)](tasks/TASK_3.13_income-scheduler-service.md) | 🔵 `backlog` | 2026-03-13 | `INCOME_SCHEDULER_INTERVAL` ENV VAR; default `1h` |
| 3.14 | [HTTP handlers: positions, income events & portfolio summary](tasks/TASK_3.14_handler-position-income.md) | 🔵 `backlog` | 2026-03-13 | `GET /v1/investments/summary`, receivable lifecycle endpoints |
| 3.15 | [Mock factory updates + Phase 3 smoke tests](tasks/TASK_3.15_smoke-tests-phase3.md) | 🔵 `backlog` | 2026-03-13 | 6 new mocks; 4 smoke test scenarios |
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
