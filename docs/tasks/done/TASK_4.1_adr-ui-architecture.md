# Task 4.1 — ADR: UI Architecture — Templ + HTMX + Alpine + Tailwind

> **Roadmap Ref:** Phase 4 — UI Foundation & Design System › Architecture Decision
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-15
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Author Architecture Decision Record 004 (`docs/ADR-004-ui-architecture.md`) documenting the rationale for the SSR-first, minimal-JavaScript UI stack: **Templ** for server-side HTML generation, **HTMX 2** for partial page updates, **Alpine.js** for lightweight client-side reactivity, and **Tailwind CSS v4** for styling. Record the WebSocket decision for real-time features. This ADR must be approved before any implementation begins.

---

## 2. Context & Motivation

Moolah is a financial SaaS with a Go backend. The frontend must be:

- **Lightweight**: load instantly on mobile connections; no multi-MB JS bundles.
- **Maintainable**: a single developer (or small team) must own both frontend and backend without context switching languages or build ecosystems.
- **Real-time capable**: balance updates, income events, and portfolio changes must push to the browser without polling.
- **Beautiful and responsive**: desktop sidebar layout + mobile drawer; dark/light theme from day one.

The ADR must evaluate alternatives (React/Next.js, SvelteKit, full HTMX without Alpine, raw HTML templates) and justify each rejection.

**Reference:** `docs/ARCHITECTURE.md` — System Overview section; `docs/ROADMAP.md` Phase 4.

---

## 3. Scope

### In scope

- [ ] Create `docs/ADR-004-ui-architecture.md` following the existing ADR format.
- [ ] Justify SSR-first approach (Go `templ` templates compiled to functions).
- [ ] Justify HTMX 2 as the hypermedia layer (partial page swaps, `hx-boost`, OOB swaps).
- [ ] Justify Alpine.js for component-local state (modals, dropdowns, form validation UI).
- [ ] Justify Tailwind CSS v4 (utility-first, zero-runtime, excellent mobile-first tooling).
- [ ] Justify WebSocket for real-time push (balance changes, income events, portfolio snapshots).
- [ ] Document the `cmd/web/` entry point and its separation from `cmd/api/`.
- [ ] Document static asset strategy: `embed.FS` for production; file-system serving in dev.
- [ ] Document build pipeline: `templ generate` → `tailwindcss` CLI → `go build`.
- [ ] Document testing strategy: `httptest` + HTML assertion (no Playwright for MVP).

### Out of scope

- Implementation of any template, component, or route (Tasks 4.2–4.9).
- Mobile app (React Native / Flutter) — deferred.

---

## 4. Technical Design

### ADR document structure

```
# ADR-004 — UI Architecture: Templ + HTMX + Alpine.js + Tailwind CSS

## Status: Accepted
## Date: 2026-03-15
## Context
## Decision
## Stack
  - Templ (a-h/templ)
  - HTMX 2
  - Alpine.js 3
  - Tailwind CSS v4
  - WebSocket (stdlib net/http upgrader)
## Consequences
  - Positive
  - Negative / Trade-offs
## Alternatives Considered
  - React / Next.js
  - SvelteKit
  - Pure server-side templates (html/template)
  - Full HTMX without Alpine
## Directory Layout
## Build Pipeline
## Testing Strategy
```

### Directory layout (to document in ADR)

```
cmd/web/
    main.go                   # web server entry point; separate binary from cmd/api/
internal/ui/
    layout/
        base.templ             # root shell (html, head, body, sidebar, topbar)
    components/
        button.templ
        input.templ
        modal.templ
        toast.templ
        table.templ
        card.templ
        badge.templ
        skeleton.templ
    pages/
        auth/
        dashboard/
        accounts/
        transactions/
        categories/
        investments/
        admin/
    middleware/
        ctx.go                 # inject render helpers into http.Request context
internal/platform/ws/
    hub.go                     # WebSocket broadcast hub
web/
    static/
        css/
            app.css            # Tailwind entrypoint (imported CSS + @apply directives)
        js/
            htmx.min.js        # vendored HTMX bundle (served via embed.FS)
            alpine.min.js      # vendored Alpine bundle
        img/
            logo.svg
```

### Files to create / modify

| Action | Path                                 | Purpose                                      |
| ------ | ------------------------------------ | -------------------------------------------- |
| CREATE | `docs/ADR-004-ui-architecture.md`    | Decision record                              |

### Key decisions to record

| Decision | Choice | Rationale |
| -------- | ------ | --------- |
| Template engine | `a-h/templ` | Type-safe, compile-checked, Go-native; superior to `html/template` for complex UIs |
| Hypermedia layer | HTMX 2 | Declarative AJAX in HTML attributes; no client-side routing needed |
| Client reactivity | Alpine.js 3 | ~15 KB; ideal for modals, dropdowns, form state; complements HTMX perfectly |
| CSS | Tailwind CSS v4 | Zero-runtime; mobile-first; replaces all custom CSS |
| Real-time | WebSocket (stdlib) | Per-tenant rooms; no external dependency; reconnect handled client-side in Alpine |
| JS delivery | Vendored via `embed.FS` | No CDN dependency; offline-friendly; version-pinned |
| Web server | Separate `cmd/web/` binary | Clean separation of concerns; can be deployed independently of the API |
| Auth sharing | Shared PASETO JWT cookie | API token reused via `HttpOnly` cookie for web sessions; same middleware |

---

## 5. Acceptance Criteria

- [x] `docs/ADR-010-ui-architecture.md` exists and is complete (all sections filled).
- [x] All four framework choices (Templ, HTMX, Alpine, Tailwind) are justified with explicit alternative comparisons.
- [x] WebSocket decision is recorded with the per-tenant room model.
- [x] Directory layout is specified in the ADR.
- [x] Build pipeline is described step-by-step.
- [x] Testing strategy (httptest + HTML assertions) is documented.
- [x] ADR is linked from `docs/ARCHITECTURE.md`.
- [x] `docs/ROADMAP.md` row 4.1 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                               |
| ---------- | ------ | ------------------------------------ |
| 2026-03-15 | —      | Task created (new)                   |
| 2026-03-15 | —      | ADR-010 authored and linked; task done |
