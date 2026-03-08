# Task 1.6.5 — Generate Swagger Documentation

> **Roadmap Ref:** Phase 1 — MVP › 1.6 Quality Gate
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-08
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement automated API documentation using Swaggo (Swagger/OpenAPI). This includes adding annotations to handlers, generating the documentation, adding a `Makefile` rule, and ensuring the documentation is verified in the CI pipeline.

---

## 2. Context & Motivation

Consistent and up-to-date API documentation is critical for frontend development and external integrations. By using Swaggo, we can generate OpenAPI specifications directly from Go code annotations, keeping docs in sync with implementation.

---

## 3. Scope

### In scope

- [ ] Add basic Swaggo annotations to `cmd/api/main.go` and existing handlers in `internal/handler/`.
- [ ] Implement `make swagger` rule to generate `docs/swagger/`.
- [ ] Add a CI step to verify that `swag init` doesn't produce uncommitted changes.
- [ ] Serve Swagger UI via a development-only route or static files.

### Out of scope

- Comprehensive documentation for all future endpoints (to be added as endpoints are implemented).

---

## 4. Technical Design

### Files to create / modify

| Action | Path | Purpose |
| ------ | ---- | ------- |
| MODIFY | `Makefile` | Add `swagger` and `swagger-check` rules |
| MODIFY | `.github/workflows/ci.yml` | Add Swagger verification step |
| MODIFY | `cmd/api/main.go` | Add General API annotations |
| MODIFY | `internal/handler/*.go` | Add operation annotations to handlers |
| CREATE | `docs/swagger/` | Generated Swagger files (swagger.json, swagger.yaml, etc.) |

---

## 5. Acceptance Criteria

- [ ] `make swagger` generates valid OpenAPI 2.0/3.0 files in `docs/swagger/`.
- [ ] Pre-existing handlers (`auth_handler.go`) have basic annotations (Summary, Tags, Responses).
- [ ] CI pipeline fails if `swag init` produces changes not committed to the repo.
- [ ] `golangci-lint` passes.

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-08 | —      | Task created from user request |
