# Task 1.5.5 — `handler/tenant_handler.go` — `GetMe`, `UpdateMe`, `InviteUser`

> **Roadmap Ref:** Phase 1 — MVP › 1.5 HTTP Handler Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement the tenant HTTP handler in `internal/handler/tenant_handler.go`. It exposes three endpoints for tenant self-service: get the current tenant's profile, update the tenant's settings, and invite a new user to the household. All operations are tenant-scoped via the `tenant_id` extracted from the PASETO token in context.

---

## 2. Context & Motivation

The `TenantService` is fully implemented (Task 1.4.2) but has no HTTP entry point. Tenants need a self-service API to manage their household profile and membership. See `docs/ARCHITECTURE.md` and roadmap item 1.5.5.

---

## 3. Scope

### In scope

- [ ] `internal/handler/tenant_handler.go` — `TenantHandler` struct + 3 HTTP handler methods.
- [ ] `GET /v1/tenants/me` — returns the current tenant's profile.
- [ ] `PATCH /v1/tenants/me` — partial update of tenant settings (name).
- [ ] `POST /v1/tenants/me/invite` — invite a user by email to the household.
- [ ] Unit tests in `internal/handler/tenant_handler_test.go`.

### Out of scope

- Admin tenant management — Task 1.5.9.
- Plan upgrade/downgrade — Phase 4 (Task 4.4).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                      | Purpose                               |
| ------ | ----------------------------------------- | ------------------------------------- |
| CREATE | `internal/handler/tenant_handler.go`      | HTTP handler for tenant endpoints     |
| CREATE | `internal/handler/tenant_handler_test.go` | Unit tests with mocked TenantService  |

### Request / Response types

```go
type UpdateTenantRequest struct {
    Name *string `json:"name" validate:"omitempty,min=2,max=100"`
}

type InviteUserRequest struct {
    Email string      `json:"email" validate:"required,email"`
    Role  domain.Role `json:"role"  validate:"required,oneof=owner member"`
}
```

### API endpoints

| Method | Path                    | Auth Required | Description                         |
| ------ | ----------------------- | ------------- | ----------------------------------- |
| GET    | `/v1/tenants/me`        | ✅ Bearer     | Get current tenant profile          |
| PATCH  | `/v1/tenants/me`        | ✅ Bearer     | Update tenant settings              |
| POST   | `/v1/tenants/me/invite` | ✅ Bearer     | Invite user to household            |

### Error cases to handle

| Scenario              | Sentinel Error           | HTTP Status |
| --------------------- | ------------------------ | ----------- |
| Not found             | `domain.ErrNotFound`     | `404`       |
| Validation failure    | —                        | `422`       |
| Conflict (duplicate)  | `domain.ErrConflict`     | `409`       |
| Forbidden             | `domain.ErrForbidden`    | `403`       |

---

## 5. Acceptance Criteria

- [ ] All 3 endpoints decode, validate, and call the service correctly.
- [ ] `tenant_id` is always sourced from context (never from the request body).
- [ ] All domain error sentinels map to the correct HTTP status codes.
- [ ] Unit tests cover all happy paths and error cases.
- [ ] Test coverage for handler ≥ 80%.
- [ ] `golangci-lint run ./...` passes with zero issues.
- [ ] `gosec ./...` passes with zero issues.
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                             | Type     | Status  |
| -------------------------------------- | -------- | ------- |
| Task 1.4.2 — `service/tenant_service`  | Upstream | ✅ done |
| Task 1.1.9 — Auth middleware           | Upstream | ✅ done |
| `domain.TenantService` interface       | Upstream | ✅ done |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/handler/tenant_handler_test.go`
- **Cases:**
  - `GetMe`: tenant found → `200 OK` with tenant JSON.
  - `GetMe`: not found → `404`.
  - `UpdateMe`: valid partial update → `200 OK`.
  - `UpdateMe`: invalid body → `422`.
  - `InviteUser`: valid email → `201 Created`.
  - `InviteUser`: duplicate email → `409`.

---

## 8. Open Questions

| # | Question                                            | Owner | Resolution |
| - | --------------------------------------------------- | ----- | ---------- |
| 1 | Does `InviteUser` send an OTP email immediately?    | —     | Yes — `TenantService.InviteUser` calls `AuthService.RequestOTP` internally. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
