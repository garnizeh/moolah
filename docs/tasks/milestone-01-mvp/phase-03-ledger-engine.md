# Task 1.3.0 — The Ledger Engine (Transactions)

> **Roadmap Ref:** Phase 3 — The Ledger Engine (Transactions)
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-22
> **Assignee:** —
> **Estimated Effort:** L

---

## 1. Summary

Build the heart of the system: the transaction ledger. This handles income, expenses, categorization, and the parsing of legacy metadata.

---

## 2. Context & Motivation

To replace spreadsheets, we must accurately track every movement of money while maintaining the metadata required to reconcile against legacy logs.

- Reference: `docs/design/003-moolah-product.md` (Section 4.1 Ledger schema)

---

## 3. Scope

### In scope

- [ ] `categories` table (hierarchical).
- [ ] `transactions` table (expected_value, actual_paid, status).
- [ ] Transaction creation service with automatic status calculation.
- [ ] Metadata capture for original parsing strings.

---

## 4. Technical Design

### Files to create / modify

| Action   | Path                                      | Purpose                       |
| -------- | ----------------------------------------- | ----------------------------- |
| CREATE   | `internal/domain/transaction.go`          | Ledger logic                  |
| CREATE   | `internal/service/ledger_service.go`      | Transaction orchestration     |
| MODIFY   | `internal/platform/db/queries/ledger.sql` | sqlc queries for financial ops|

---

## 5. Acceptance Criteria

- [ ] Support for transactions with multi-line metadata (JSONB).
- [ ] Unit tests cover partial payment scenarios.
- [ ] Integration tests verify atomicity of balance updates.
