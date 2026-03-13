# Task 2.4 — Repository: `MasterPurchaseRepository` Implementation + Integration Tests

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › Repository Layer
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-12
> **Assignee:** GitHub Copilot

---

## 1. Summary
Implement the concrete `MasterPurchaseRepository` using `sqlc` generated queries.

---

## 5. Acceptance Criteria
- [x] Implements `domain.MasterPurchaseRepository` interface.
- [x] Includes `WHERE tenant_id = $1` and `AND deleted_at IS NULL` (via sqlc).
- [x] 100% test coverage with integration tests.
- [x] Translation of pgx error to domain sentinel errors.

---

## 8. Change Log
| Date       | Author         | Change |
| ---------- | -------------- | ------ |
| 2026-03-12 | —              | Task created from roadmap |
| 2026-03-12 | GitHub Copilot | Task completed: Implemented repository with full CRUD and integration tests |
