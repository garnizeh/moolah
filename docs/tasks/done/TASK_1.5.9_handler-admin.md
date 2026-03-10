# Task 1.5.9 — `handler/admin_handler.go` — sysadmin routes

> **Roadmap Ref:** Phase 1 — MVP › 1.5 HTTP Handler Layer
> **Status:** � `in-progress`
> **Last Updated:** 2026-03-09
> **Assignee:** GitHub Copilot
> **Estimated Effort:** M

---

## 1. Summary

Implement the admin HTTP handler in `internal/handler/admin_handler.go`. It exposes cross-tenant sysadmin routes for managing all tenants, users, and viewing global audit logs. All routes require the `sysadmin` role enforced by the `RequireRole` middleware.

---

## 2. Context & Motivation

The `AdminService` is fully implemented (Task 1.4.6), but has no HTTP entry point. These routes bypass standard tenant isolation and must be carefully protected. The handler implements the elevated-privilege API surface for system operators. See `docs/ARCHITECTURE.md` and roadmap item 1.5.9.

---

## 3. Scope

### In scope

- [ ] `internal/handler/admin_handler.go` — `AdminHandler` struct + 10 HTTP handler methods.
- [ ] `GET /v1/admin/tenants` — list all tenants, optional `?with_deleted=true`.
- [ ] `GET /v1/admin/tenants/{id}` — get a tenant by ID.
- [ ] `PATCH /v1/admin/tenants/{id}/plan` — update tenant plan.
- [ ] `POST /v1/admin/tenants/{id}/suspend` — suspend a tenant.
- [ ] `POST /v1/admin/tenants/{id}/restore` — restore a suspended tenant.
- [ ] `DELETE /v1/admin/tenants/{id}` — hard delete (requires `X-Confirm-Token: {id}` header).
- [ ] `GET /v1/admin/users` — list all users.
- [ ] `GET /v1/admin/users/{id}` — get a user by ID.
- [ ] `DELETE /v1/admin/users/{id}` — force-delete a user.
- [ ] `GET /v1/admin/audit-logs` — global audit log with filters.
- [ ] Unit tests in `internal/handler/admin_handler_test.go`.

### Out of scope

- Billing / MRR analytics — Phase 4 (Task 4.6).
- Plan quota enforcement — Phase 4 (Task 4.2).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                      | Purpose                              |
| ------ | ----------------------------------------- | ------------------------------------ |
| CREATE | `internal/handler/admin_handler.go`       | HTTP handler for admin endpoints     |
| CREATE | `internal/handler/admin_handler_test.go`  | Unit tests with mocked AdminService  |

### Request / Response types

```go
type UpdateTenantPlanRequest struct {
    Plan domain.TenantPlan `json:"plan" validate:"required"`
}
```

### Hard-delete confirmation

`HardDeleteTenant` reads the `X-Confirm-Token` header. This value must equal the tenant ID. If it does not match, the handler returns `400 Bad Request` before even calling the service. The service performs its own identical check as a defence-in-depth guard.

### API endpoints

| Method | Path                                   | Auth Required | Role       | Description                |
| ------ | -------------------------------------- | ------------- | ---------- | -------------------------- |
| GET    | `/v1/admin/tenants`                    | ✅ Bearer     | `sysadmin` | List all tenants           |
| GET    | `/v1/admin/tenants/{id}`               | ✅ Bearer     | `sysadmin` | Get tenant by ID           |
| PATCH  | `/v1/admin/tenants/{id}/plan`          | ✅ Bearer     | `sysadmin` | Update tenant plan         |
| POST   | `/v1/admin/tenants/{id}/suspend`       | ✅ Bearer     | `sysadmin` | Suspend tenant             |
| POST   | `/v1/admin/tenants/{id}/restore`       | ✅ Bearer     | `sysadmin` | Restore tenant             |
| DELETE | `/v1/admin/tenants/{id}`               | ✅ Bearer     | `sysadmin` | Hard-delete tenant         |
| GET    | `/v1/admin/users`                      | ✅ Bearer     | `sysadmin` | List all users             |
| GET    | `/v1/admin/users/{id}`                 | ✅ Bearer     | `sysadmin` | Get user by ID             |
| DELETE | `/v1/admin/users/{id}`                 | ✅ Bearer     | `sysadmin` | Force-delete user          |
| GET    | `/v1/admin/audit-logs`                 | ✅ Bearer     | `sysadmin` | Global audit log list      |

### Error cases to handle

| Scenario                     | Sentinel Error           | HTTP Status |
| ---------------------------- | ------------------------ | ----------- |
| Not found                    | `domain.ErrNotFound`     | `404`       |
| Invalid confirmation token   | `domain.ErrInvalidInput` | `400`       |
| Validation failure           | —                        | `422`       |
| Insufficient role            | `domain.ErrForbidden`    | `403`       |

---

## 5. Acceptance Criteria

- [ ] All 10 endpoints decode, validate, and call the service correctly.
- [ ] `HardDeleteTenant` validates `X-Confirm-Token` header before forwarding.
- [ ] All routes require `sysadmin` role via `RequireRole` middleware in `routes.go`.
- [ ] All domain error sentinels map to the correct HTTP status codes.
- [ ] Unit tests cover all happy paths and all error cases.
- [ ] Test coverage for handler ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                            | Type     | Status  |
| ------------------------------------- | -------- | ------- |
| Task 1.4.6 — `service/admin_service`  | Upstream | ✅ done |
| Task 1.1.9 — Auth middleware          | Upstream | ✅ done |
| `domain.AdminService` interface       | Upstream | ✅ done |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/handler/admin_handler_test.go`
- **Cases:**
  - `ListTenants`: returns all → `200 OK`.
  - `GetTenant`: found → `200 OK`.
  - `GetTenant`: not found → `404`.
  - `UpdateTenantPlan`: valid → `200 OK`.
  - `SuspendTenant`: success → `204 No Content`.
  - `HardDeleteTenant`: missing header → `400`.
  - `HardDeleteTenant`: wrong token → `400`.
  - `HardDeleteTenant`: correct token → `204 No Content`.
  - `ForceDeleteUser`: success → `204 No Content`.
  - `ListAuditLogs`: returns logs → `200 OK`.

---

## 8. Open Questions

| # | Question                                                          | Owner | Resolution |
| - | ----------------------------------------------------------------- | ----- | ---------- |
| 1 | Should `HardDeleteTenant` return `204` or `200` with a message?   | —     | `204 No Content`. |
| 2 | Should audit log list support pagination?                         | —     | Yes — `limit`/`offset` query params passed to `ListAuditLogsParams`. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
