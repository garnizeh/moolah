# Task 1.5.3 — `cmd/api/routes.go` — all route registrations

> **Roadmap Ref:** Phase 1 — MVP › 1.5 HTTP Handler Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement `cmd/api/routes.go`, which registers all HTTP routes onto an `http.ServeMux` using Go 1.22+ `METHOD /path/{param}` routing syntax. It wires each handler to its path, applies per-route middleware (auth, rate-limit, idempotency), and returns the populated mux.

---

## 2. Context & Motivation

Centralising all routes in one file makes the API surface easy to audit and change. The Go 1.22 stdlib mux eliminates any need for an external router. Per the project architecture, no external router framework is allowed. See `docs/ARCHITECTURE.md` and roadmap item 1.5.3.

---

## 3. Scope

### In scope

- [ ] `cmd/api/routes.go` — `NewRouter(...)` function returning `*http.ServeMux`.
- [ ] Register all Phase 1 routes grouped by domain.
- [ ] Apply per-route middleware using handler wrapping: `RequireAuth`, `RequireRole`, `RateLimit`, `Idempotency`.
- [ ] Use Go 1.22 pattern syntax: `mux.Handle("POST /v1/auth/otp/request", handler)`.

### Out of scope

- Handler implementations — Tasks 1.5.4–1.5.9.
- Swagger/OpenAPI spec — Task 1.5.10.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                 | Purpose                          |
| ------ | -------------------- | -------------------------------- |
| CREATE | `cmd/api/routes.go`  | Centralised route registration   |

### Route table

| Method | Path                              | Handler                    | Middleware                       |
| ------ | --------------------------------- | -------------------------- | -------------------------------- |
| POST   | `/v1/auth/otp/request`            | `AuthHandler.RequestOTP`   | RateLimit                        |
| POST   | `/v1/auth/otp/verify`             | `AuthHandler.VerifyOTP`    | RateLimit                        |
| POST   | `/v1/auth/token/refresh`          | `AuthHandler.RefreshToken` | RequireAuth                      |
| GET    | `/v1/tenants/me`                  | `TenantHandler.GetMe`      | RequireAuth                      |
| PATCH  | `/v1/tenants/me`                  | `TenantHandler.UpdateMe`   | RequireAuth, Idempotency         |
| POST   | `/v1/tenants/me/invite`           | `TenantHandler.InviteUser` | RequireAuth, Idempotency         |
| GET    | `/v1/accounts`                    | `AccountHandler.List`      | RequireAuth                      |
| POST   | `/v1/accounts`                    | `AccountHandler.Create`    | RequireAuth, Idempotency         |
| GET    | `/v1/accounts/{id}`               | `AccountHandler.GetByID`   | RequireAuth                      |
| PATCH  | `/v1/accounts/{id}`               | `AccountHandler.Update`    | RequireAuth, Idempotency         |
| DELETE | `/v1/accounts/{id}`               | `AccountHandler.Delete`    | RequireAuth                      |
| GET    | `/v1/categories`                  | `CategoryHandler.List`     | RequireAuth                      |
| POST   | `/v1/categories`                  | `CategoryHandler.Create`   | RequireAuth, Idempotency         |
| GET    | `/v1/categories/{id}`             | `CategoryHandler.GetByID`  | RequireAuth                      |
| PATCH  | `/v1/categories/{id}`             | `CategoryHandler.Update`   | RequireAuth, Idempotency         |
| DELETE | `/v1/categories/{id}`             | `CategoryHandler.Delete`   | RequireAuth                      |
| GET    | `/v1/transactions`                | `TransactionHandler.List`  | RequireAuth                      |
| POST   | `/v1/transactions`                | `TransactionHandler.Create`| RequireAuth, Idempotency         |
| GET    | `/v1/transactions/{id}`           | `TransactionHandler.GetByID`| RequireAuth                     |
| PATCH  | `/v1/transactions/{id}`           | `TransactionHandler.Update`| RequireAuth, Idempotency         |
| DELETE | `/v1/transactions/{id}`           | `TransactionHandler.Delete`| RequireAuth                      |
| GET    | `/v1/admin/tenants`               | `AdminHandler.ListTenants` | RequireAuth, RequireRole(sysadmin)|
| GET    | `/v1/admin/tenants/{id}`          | `AdminHandler.GetTenant`   | RequireAuth, RequireRole(sysadmin)|
| PATCH  | `/v1/admin/tenants/{id}/plan`     | `AdminHandler.UpdatePlan`  | RequireAuth, RequireRole(sysadmin)|
| POST   | `/v1/admin/tenants/{id}/suspend`  | `AdminHandler.Suspend`     | RequireAuth, RequireRole(sysadmin)|
| POST   | `/v1/admin/tenants/{id}/restore`  | `AdminHandler.Restore`     | RequireAuth, RequireRole(sysadmin)|
| DELETE | `/v1/admin/tenants/{id}`          | `AdminHandler.HardDelete`  | RequireAuth, RequireRole(sysadmin)|
| GET    | `/v1/admin/users`                 | `AdminHandler.ListUsers`   | RequireAuth, RequireRole(sysadmin)|
| DELETE | `/v1/admin/users/{id}`            | `AdminHandler.ForceDeleteUser`| RequireAuth, RequireRole(sysadmin)|
| GET    | `/v1/admin/audit-logs`            | `AdminHandler.ListAuditLogs`| RequireAuth, RequireRole(sysadmin)|

---

## 5. Acceptance Criteria

- [ ] All routes listed in the route table above are registered.
- [ ] Go 1.22 `METHOD /path/{param}` syntax is used throughout.
- [ ] All mutating `POST`/`PATCH`/`DELETE` routes use the appropriate middleware.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                    | Type     | Status     |
| --------------------------------------------- | -------- | ---------- |
| Task 1.5.4 — `handler/auth_handler.go`        | Upstream | 🔵 backlog |
| Task 1.5.5 — `handler/tenant_handler.go`      | Upstream | 🔵 backlog |
| Task 1.5.6 — `handler/account_handler.go`     | Upstream | 🔵 backlog |
| Task 1.5.7 — `handler/category_handler.go`    | Upstream | 🔵 backlog |
| Task 1.5.8 — `handler/transaction_handler.go` | Upstream | 🔵 backlog |
| Task 1.5.9 — `handler/admin_handler.go`       | Upstream | 🔵 backlog |
| Task 1.1.9 — Auth middleware                   | Upstream | ✅ done    |
| Task 1.1.10 — Rate-limit middleware            | Upstream | ✅ done    |
| Task 1.5.11 — Idempotency wiring              | Related  | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests

N/A — route registration is validated by smoke tests.

### Integration / smoke test

- Task 1.6.4 Postman/httpie collection exercises every registered route.

---

## 8. Open Questions

| # | Question                                                    | Owner | Resolution |
| - | ----------------------------------------------------------- | ----- | ---------- |
| 1 | Version prefix `/v1` hardcoded or from config?              | —     | Hardcoded for Phase 1. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
