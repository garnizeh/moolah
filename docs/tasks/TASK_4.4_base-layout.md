# Task 4.4 — Base Layout Template

> **Roadmap Ref:** Phase 4 — UI Foundation & Design System › Layout
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-15
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the responsive application shell in `internal/ui/layout/base.templ`: the `<html>` document wrapper, the sticky topbar, the responsive collapsible sidebar (desktop persistent, mobile overlay drawer), the main content slot, and the global notification area (toasts). Every page in Phases 5 and 6 will compose this layout.

---

## 2. Context & Motivation

All authenticated pages share the same outer shell. Building it once — correctly and responsively — means every subsequent task only needs to fill in the `content` slot. The layout must work perfectly on a 320px-wide phone screen and on a 2560px widescreen monitor without JavaScript-dependent layout shifts.

**Mobile strategy:**

- **Desktop (≥ lg):** persistent left sidebar (256px wide), content fills remaining width.
- **Tablet/Mobile (< lg):** sidebar hidden by default; hamburger button in topbar opens it as an overlay drawer.

**Depends on:** Task 4.2 (toolchain), Task 4.3 (design tokens).

---

## 3. Scope

### In scope

- [ ] `internal/ui/layout/base.templ` — root document shell with `content templ.Component` parameter.
- [ ] `internal/ui/layout/sidebar.templ` — navigation sidebar with grouped links and active-route highlighting.
- [ ] `internal/ui/layout/topbar.templ` — sticky top bar; hamburger (mobile), page title slot, user menu dropdown, theme toggle.
- [ ] `internal/ui/layout/nav_item.templ` — reusable navigation link component (icon + label + active state).
- [ ] Responsive sidebar behaviour: Alpine.js `open` store drives visibility; overlay closes on outside click or `Escape`.
- [ ] Active route detection: current URL path passed from handler and compared against nav item `href`.
- [ ] User menu dropdown in topbar: displays user email, links to tenant settings and logout.
- [ ] Dark/light theme toggle button in topbar (calls `$store.theme.toggle()`).
- [ ] HTMX progress bar indicator wired to `htmx:beforeRequest` / `htmx:afterRequest` events.
- [ ] `<meta>` tags: charset, viewport, CSRF token (for non-idempotent HTMX requests), Open Graph basics.
- [ ] Favicon and web app manifest (`/static/img/logo.svg`).
- [ ] Unit test: render the base layout and assert key structural HTML elements are present.

### Out of scope

- Page-specific content (each page task creates its own templ component).
- Footer (deferred — not needed in MVP dashboard layout).
- Breadcrumb component (deferred to Phase 5).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                   | Purpose                                           |
| ------ | -------------------------------------- | ------------------------------------------------- |
| CREATE | `internal/ui/layout/base.templ`        | Root document shell                               |
| CREATE | `internal/ui/layout/sidebar.templ`     | Left navigation sidebar                           |
| CREATE | `internal/ui/layout/topbar.templ`      | Top navigation bar                                |
| CREATE | `internal/ui/layout/nav_item.templ`    | Reusable nav link                                 |
| CREATE | `internal/ui/layout/layout_test.go`    | Unit tests: render and assert HTML structure      |

### Layout parameter type

```go
// internal/ui/layout/base.templ

package layout

type BaseProps struct {
    Title       string              // <title> tag value (appended with " — Moolah")
    CurrentPath string              // active route path for nav highlighting
    User        *domain.User        // logged-in user (nil for auth pages)
    Tenant      *domain.Tenant      // active tenant (nil for auth pages)
    Content     templ.Component     // page body injected into main slot
}
```

### Template structure (pseudo-code)

```
base(props BaseProps)
  <html lang="en" x-data x-bind:class="{dark: $store.theme.dark}">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>{ props.Title } — Moolah</title>
      <link rel="stylesheet" href="/static/css/app.min.css">
      <script defer src="/static/js/alpine.min.js"></script>
      <script defer src="/static/js/htmx.min.js"></script>
    </head>
    <body class="bg-[--color-bg] text-[--color-text-primary]">
      <!-- Mobile sidebar overlay backdrop -->
      <div x-show="$store.nav.open" @click="$store.nav.close()"
           class="fixed inset-0 z-20 bg-black/50 lg:hidden" .../>

      <!-- Sidebar -->
      @sidebar(props)

      <!-- Main column -->
      <div class="lg:pl-64 flex flex-col min-h-screen">
        @topbar(props)

        <!-- HTMX progress bar -->
        <div id="htmx-progress" .../>

        <!-- Page content slot -->
        <main id="main-content" class="flex-1 p-4 lg:p-8">
          @props.Content
        </main>
      </div>

      <!-- Global toast container (Alpine.js $store.toasts) -->
      @toastContainer()
    </body>
  </html>
```

### Alpine.js stores (injected in `<head>`)

```javascript
Alpine.store('nav', {
    open: false,
    toggle() { this.open = !this.open; },
    close() { this.open = false; }
});

Alpine.store('theme', {
    dark: localStorage.getItem('theme') === 'dark' || false,
    toggle() {
        this.dark = !this.dark;
        localStorage.setItem('theme', this.dark ? 'dark' : 'light');
    }
});

Alpine.store('toasts', {
    items: [],
    add(msg, type = 'info', duration = 4000) {
        const id = Date.now();
        this.items.push({ id, msg, type });
        setTimeout(() => this.remove(id), duration);
    },
    remove(id) { this.items = this.items.filter(t => t.id !== id); }
});
```

### Navigation structure

| Section | Items |
| ------- | ----- |
| Main | Dashboard, Accounts, Transactions |
| Planning | Categories, Master Purchases |
| Investments | Portfolio, Positions, Income |
| Admin | Tenants, Users, Audit Logs *(sysadmin only)* |

---

## 5. Acceptance Criteria

- [ ] `make templ` compiles `base.templ`, `sidebar.templ`, `topbar.templ`, `nav_item.templ` without errors.
- [ ] Rendered HTML passes W3C structural validation (no unclosed tags, correct nesting).
- [ ] Sidebar collapses to hidden on `< lg` screen width; hamburger shows.
- [ ] Sidebar opens as overlay when hamburger is clicked; closes on backdrop click and `Escape`.
- [ ] Active nav item is visually highlighted for the current route.
- [ ] Dark mode toggle persists to `localStorage` and applies `.dark` class to `<html>`.
- [ ] Theme respects `prefers-color-scheme` on first visit when no `localStorage` value is set.
- [ ] User menu dropdown shows user email and links to settings / logout.
- [ ] HTMX progress bar appears during HTMX requests.
- [ ] `layout_test.go` asserts: `<title>`, `<main id="main-content">`, sidebar link to `/dashboard`, theme toggle button.
- [ ] `golangci-lint run ./internal/ui/...` passes.
- [ ] `docs/ROADMAP.md` row 4.4 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change             |
| ---------- | ------ | ------------------ |
| 2026-03-15 | —      | Task created (new) |
