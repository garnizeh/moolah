# Task 1.5.10 — Swaggo annotations on all handlers; `swag init` verified in CI

> **Roadmap Ref:** Phase 1 — MVP › 1.5 HTTP Handler Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-07
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Add Swaggo (`github.com/swaggo/swag`) doc comments to all Phase 1 HTTP handlers and verify that `swag init` is idempotent and committed in CI. The generated `docs/swagger/` output (OpenAPI 2.0 spec) must be kept in sync with handler annotations at every PR.

---

## 2. Context & Motivation

The API has functional handlers but no machine-readable documentation. Swaggo generates an OpenAPI spec from structured Go comments, enabling client generation, Postman import, and developer self-service. This task annotates all handlers completed in Tasks 1.5.4–1.5.9. See roadmap item 1.5.10.

---

## 3. Scope

### In scope

- [ ] Add `@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router` annotations to every handler method in:
  - `internal/handler/auth_handler.go`
  - `internal/handler/tenant_handler.go`
  - `internal/handler/account_handler.go`
  - `internal/handler/category_handler.go`
  - `internal/handler/transaction_handler.go`
  - `internal/handler/admin_handler.go`
- [ ] Top-level `@title`, `@version`, `@host`, `@BasePath`, `@securityDefinitions` in `cmd/api/main.go` or a dedicated `docs.go`.
- [ ] Run `swag init` and commit the generated `docs/swagger/` directory.
- [ ] Add `make swagger` target to `Makefile`.
- [ ] Add `swag init --output docs/swagger && git diff --exit-code docs/swagger` check to GitHub Actions CI.

### Out of scope

- Swagger UI serving endpoint — may be added in Phase 5 observability work.
- OpenAPI 3.0 conversion — Swaggo generates 2.0; migration deferred.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                   | Purpose                              |
| ------ | -------------------------------------- | ------------------------------------ |
| MODIFY | All `internal/handler/*.go` files      | Add Swaggo doc comments              |
| CREATE | `cmd/api/docs.go` (or inline in main)  | Top-level API metadata comment       |
| CREATE | `docs/swagger/swagger.json`            | Generated OpenAPI 2.0 spec           |
| CREATE | `docs/swagger/swagger.yaml`            | Generated YAML spec                  |
| MODIFY | `Makefile`                             | Add `swagger` target                 |
| MODIFY | `.github/workflows/ci.yml`             | Add swagger drift check step         |

### Example annotation

```go
// RequestOTP sends a one-time password to the given email address.
// @Summary     Request OTP
// @Description Send a 6-digit OTP to the given email for authentication.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body RequestOTPRequest true "Email address"
// @Success     202 {object} map[string]string
// @Failure     422 {object} ErrorResponse
// @Failure     429 {object} ErrorResponse "Rate limit exceeded"
// @Router      /v1/auth/otp/request [post]
func (h *AuthHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
```

### Security definition

```go
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
```

---

## 5. Acceptance Criteria

- [ ] Every handler method has complete Swaggo annotations.
- [ ] `swag init` produces a valid `docs/swagger/swagger.json` without errors.
- [ ] `git diff --exit-code docs/swagger/` passes in CI (no uncommitted drift).
- [ ] `make swagger` target works locally.
- [ ] `golangci-lint run ./...` passes with zero issues (swagger output excluded from lint via `.golangci.yml`).
- [ ] `docs/ROADMAP.md` row updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                               | Type     | Status     |
| ---------------------------------------- | -------- | ---------- |
| Task 1.5.4 — `handler/auth_handler`      | Upstream | 🔵 backlog |
| Task 1.5.5 — `handler/tenant_handler`    | Upstream | 🔵 backlog |
| Task 1.5.6 — `handler/account_handler`   | Upstream | 🔵 backlog |
| Task 1.5.7 — `handler/category_handler`  | Upstream | 🔵 backlog |
| Task 1.5.8 — `handler/transaction_handler`| Upstream | 🔵 backlog |
| Task 1.5.9 — `handler/admin_handler`     | Upstream | 🔵 backlog |

---

## 7. Testing Plan

### Unit tests

N/A — Swaggo annotations are documentation, not logic.

### CI check

- `swag init` must exit 0.
- Spec file must not have uncommitted diff in CI.

---

## 8. Open Questions

| # | Question                                               | Owner | Resolution |
| - | ------------------------------------------------------ | ----- | ---------- |
| 1 | Serve Swagger UI at `/swagger/` in dev mode?           | —     | Deferred to Phase 5. Spec is available as static file only. |
| 2 | Should `docs/swagger/` be `.gitignore`d or committed?  | —     | Committed — drift detection in CI requires it to be in source control. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-07 | —      | Task created from roadmap |
