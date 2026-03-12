# Task 2.9 — Remainder-Cent Handling in Instalment Calculation

> **Roadmap Ref:** Phase 2 — Credit Card & Installment Tracking › Service Layer
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-12
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Ensure that the sum of all projected instalment amounts **always exactly equals** `TotalAmountCents`, even when the total is not evenly divisible by the number of instalments. The final instalment absorbs the integer-division remainder. This logic lives inside `MasterPurchaseService.ProjectInstallments` (introduced in Task 2.5) and is validated by dedicated unit tests here.

---

## 2. Context & Motivation

Financial systems must never lose or create money through rounding. A purchase of R$ 100.01 divided into 3 instalments yields: `100 + 100 + 101` cents (base=33, remainder=1, last instalment = 33+1 = 34... wait that's only 100, we need 33+33+35... Actually: 100/3=33 r1, so [33, 33, 34] = 100 ✓).

Using `float64` for this would introduce floating-point drift. The `int64` integer-division approach is the only correct method. This task documents the algorithm, adds property-based-style exhaustive unit tests, and ensures the implementation in Task 2.5 satisfies the invariant.

---

## 3. Scope

### In scope

- [ ] Document the remainder-cent algorithm in code comments within `master_purchase_service.go`.
- [ ] Dedicated unit test file (or section) with table-driven tests covering edge cases.
- [ ] Verify the invariant: `sum(ProjectedInstallments.AmountCents) == mp.TotalAmountCents` for all valid inputs.

### Out of scope

- Changing the algorithm location — it remains in `MasterPurchaseService.ProjectInstallments` (Task 2.5).
- Any DB changes.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                                  | Purpose                                          |
| ------ | ----------------------------------------------------- | ------------------------------------------------ |
| MODIFY | `internal/service/master_purchase_service.go`         | Add algorithm doc comment to `ProjectInstallments` |
| MODIFY | `internal/service/master_purchase_service_test.go`    | Add exhaustive remainder-cent test cases         |

### Algorithm (canonical reference)

```go
// ProjectInstallments computes each instalment's due date and amount.
//
// Remainder-cent rule: when TotalAmountCents is not evenly divisible by
// InstallmentCount, the LAST instalment absorbs all remaining cents so that:
//
//   sum(result[i].AmountCents for i in 0..N-1) == mp.TotalAmountCents
//
// Example: total=1000, count=3 → base=333, remainder=1 → [333, 333, 334]
// Example: total=1001, count=3 → base=333, remainder=2 → [333, 333, 335]
// Example: total=1200, count=3 → base=400, remainder=0 → [400, 400, 400]
func (s *masterPurchaseService) ProjectInstallments(mp *domain.MasterPurchase) []domain.ProjectedInstallment {
    base      := mp.TotalAmountCents / int64(mp.InstallmentCount)
    remainder := mp.TotalAmountCents % int64(mp.InstallmentCount)

    result := make([]domain.ProjectedInstallment, mp.InstallmentCount)
    for i := range result {
        amount := base
        if i == int(mp.InstallmentCount)-1 {
            amount += remainder
        }
        result[i] = domain.ProjectedInstallment{
            InstallmentNumber: int32(i + 1),
            DueDate:           mp.FirstInstallmentDate.AddDate(0, i, 0),
            AmountCents:       amount,
        }
    }
    return result
}
```

### Test cases (table-driven)

| TotalAmountCents | InstallmentCount | Expected Distribution                 | Sum Check |
| ---------------- | ---------------- | ------------------------------------- | --------- |
| 1200             | 3                | [400, 400, 400]                       | ✅ 1200   |
| 1000             | 3                | [333, 333, 334]                       | ✅ 1000   |
| 1001             | 3                | [333, 333, 335]                       | ✅ 1001   |
| 1                | 2                | [0, 1]                                | ✅ 1      |
| 100              | 1                | N/A (min=2, rejected at validation)   | —         |
| 99               | 2                | [49, 50]                              | ✅ 99     |
| 10000            | 12               | [833, 833, 833, …×11, 837]            | ✅ 10000  |
| 1                | 48               | [0×47, 1]                             | ✅ 1      |

---

## 5. Acceptance Criteria

- [ ] `ProjectInstallments` has a doc comment explaining the remainder-cent rule with examples.
- [ ] Unit tests cover all table cases above.
- [ ] Invariant test: for all entries, `sum(AmountCents) == TotalAmountCents`.
- [ ] No `float64` appears anywhere in the instalment calculation path.
- [ ] Edge case: `remainder == 0` → all instalments equal, last is not inflated.
- [ ] `golangci-lint run ./...` passes.
- [ ] `gosec ./...` passes.
- [ ] `docs/ROADMAP.md` row 2.9 updated to ✅ `done`.

---

## 6. Dependencies

| Dependency                                          | Type     | Status       |
| --------------------------------------------------- | -------- | ------------ |
| Task 2.5 — `ProjectInstallments` implementation    | Upstream | 🔵 backlog   |

---

## 7. Testing Plan

### Unit tests (`_test.go`, no build tag)

- **File:** `internal/service/master_purchase_service_test.go` (extended section)
- **Pattern:** Table-driven tests with `require.Equal(t, mp.TotalAmountCents, sumAmounts(instalments))`.
- **Helper:** `sumAmounts([]domain.ProjectedInstallment) int64` — local test helper.

### Integration tests (`//go:build integration`)

N/A — pure arithmetic function; no DB dependency.

---

## 8. Open Questions

| # | Question                                                                              | Owner | Resolution |
| - | ------------------------------------------------------------------------------------- | ----- | ---------- |
| 1 | Should remainder go to the FIRST or LAST instalment?                                  | —     | LAST — matches common credit card issuer behaviour (last bill is slightly higher). |
| 2 | What if `TotalAmountCents < InstallmentCount` (e.g., 1 cent into 3 instalments)?     | —     | `base=0`, `remainder=1`; first N-1 instalments are R$0.00, last carries the total. Allowed — validation only enforces `TotalAmountCents > 0`. |

---

## 9. Change Log

| Date       | Author | Change                    |
| ---------- | ------ | ------------------------- |
| 2026-03-12 | —      | Task created from roadmap |
