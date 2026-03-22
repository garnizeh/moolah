# Task 1.5.0 — Hypermedia UI (HTMX/Tailwind)

> **Roadmap Ref:** Phase 5 — Hypermedia UI (HTMX/Tailwind)
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-22
> **Assignee:** —
> **Estimated Effort:** L

---

## 1. Summary

Develop the user-facing web interface using HTMX for a dynamic feel with minimal JavaScript. Focus on a premium, dashboard-centric design.

---

## 2. Context & Motivation

Moolah's success depends on it being easier and more satisfying to use than a spreadsheet. HTMX allows for high interactivity while keeping the business logic on the Go server.

---

## 3. Scope

### In scope

- [ ] Base layout with Tailwind CSS premium dark mode.
- [ ] Net Worth and Account Overview dashboards.
- [ ] Transaction entry forms with real-time feedback (HTMX).
- [ ] Variance Analysis view.

---

## 4. Technical Design

| Action   | Path                                      | Purpose                       |
| -------- | ----------------------------------------- | ----------------------------- |
| CREATE   | `internal/api/handler/dashboard.go`       | Render views                  |
| CREATE   | `web/templates/layouts/base.html`         | Shared skeleton               |
| CREATE   | `web/templates/components/`               | Reusable UI fragments         |

---

## 5. Acceptance Criteria

- [ ] Mobile-responsive layout.
- [ ] Navigating between pages does not trigger a full page reload (HTMX).
- [ ] Dashboard correctly sums balances across entities and accounts.
