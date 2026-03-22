# Task 1.2.0 — Domain Modeling & Core Registry

> **Roadmap Ref:** Phase 2 — Domain Modeling & Core Registry
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-22
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the fundamental entities of the system: Currencies, Entities (family members/centers), and Accounts. These form the registry that all transactions will reference.

---

## 2. Context & Motivation

Moolah requires strict multi-entity and multi-currency support. We need to store monetary values in cents and handle display precision via the Currency configuration.

- Reference: `docs/design/003-moolah-product.md` (Section 4.1 ER Diagram)
- Link to relevant roadmap row: `docs/tasks/roadmap.md#Phase-2`

---

## 3. Scope

### In scope

- [x] `currencies` table (code, symbol, decimals).
- [x] `entities` table (name, role, metadata JSONB).
- [x] `accounts` table (entity_id, currency_id, balance_cents).
- [x] CRUD services for these registries correctly handling ULID IDs.

---

## 4. Technical Design

### Files to create / modify

| Action   | Path                                      | Purpose                       |
| -------- | ----------------------------------------- | ----------------------------- |
| CREATE   | `internal/domain/currency.go`             | Model + Logic for decimals    |
| CREATE   | `internal/domain/entity.go`               | Entity definition             |
| CREATE   | `internal/platform/db/migrations/001_initial.sql` | Schema creation |

---

## 5. Acceptance Criteria

- [x] Repository interface is defined in `internal/domain/`.
- [x] Every SQL query includes `WHERE tenant_id` (if applicable) and JSONB tags.
- [x] Cents precision is strictly enforced in the database.

