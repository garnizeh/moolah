# Task 1.4.0 — Obligations & Installments

> **Roadmap Ref:** Phase 4 — Obligations & Installments
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-22
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement a specific module for long-term financial obligations (loans, financing, recurring bills) that automatically generates future ledger entries.

---

## 2. Context & Motivation

Legacy spreadsheets tracked items like "Volkswagen 11/36". The system should automate this by creating the full payment schedule as "Expected" transactions.

---

## 3. Scope

### In scope

- [ ] `long_term_obligations` table and model.
- [ ] Installment generation logic (bulk creation of transactions).
- [ ] Contract grouping for related installments.

---

## 4. Technical Design

### Key interfaces / types

```go
type Obligation struct {
    ID                string
    TotalInstallments int
    CurrentInstallment int
    ContractID        string
}
```

---

## 5. Acceptance Criteria

- [ ] Creating an obligation with 12 installments generates 12 PENDING transactions.
- [ ] Payments correctly update the remaining principal of the contract.
