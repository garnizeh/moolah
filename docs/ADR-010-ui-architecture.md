# ADR-010 — UI Architecture: Templ + HTMX + Alpine.js + Tailwind CSS v4

> **Status:** Accepted
> **Date:** 2026-03-15
> **Deciders:** Engineering Team
> **Phase:** 4 — UI Foundation & Design System
> **Supersedes:** —

---

## Table of Contents

1. [Context](#1-context)
2. [Decision](#2-decision)
3. [Stack Overview](#3-stack-overview)
4. [Directory Layout](#4-directory-layout)
5. [Binary Separation: `cmd/web/` vs `cmd/api/`](#5-binary-separation-cmdweb-vs-cmdapi)
6. [Authentication Strategy](#6-authentication-strategy)
7. [Real-Time: WebSocket Architecture](#7-real-time-websocket-architecture)
8. [Static Asset Strategy](#8-static-asset-strategy)
9. [Build Pipeline](#9-build-pipeline)
10. [Testing Strategy](#10-testing-strategy)
11. [Security Considerations](#11-security-considerations)
12. [Consequences](#12-consequences)
13. [Alternatives Considered](#13-alternatives-considered)
14. [Open Questions](#14-open-questions)

---

## 1. Context

Moolah's API (Phases 1–3) is complete. The platform needs a production-quality web UI to allow household members to:

- View and manage accounts, transactions, categories, and installment purchases.
- Track investment positions, portfolio snapshots, and income receivables.
- Receive real-time updates when balances change, income is received, or a portfolio snapshot completes.
- Operate comfortably on both desktop and mobile devices.
- Switch between light and dark visual themes.

### Constraints

| Constraint | Detail |
| --- | --- |
| **Team size** | Small (1–2 engineers). The same people own both the Go backend and the frontend. |
| **Language boundary** | Every context switch away from Go (e.g. TypeScript, Node.js toolchains) is friction and a maintenance surface. |
| **Bundle size** | The target audience includes mobile users on variable connections. Sub-50 KB total JS budget. |
| **Runtime complexity** | No client-side routing, hydration mismatch errors, or serialisation round-trips. The server owns state. |
| **Real-time** | Balance changes and income events must be pushed to the browser with latency < 2 s without polling. |
| **Build pipeline** | Must integrate cleanly with existing `Makefile`, Docker build, and GitHub Actions CI pipeline. |
| **Security** | OWASP Top 10 compliance is non-negotiable; XSS, CSRF, cookie theft, and WebSocket hijacking must be explicitly addressed. |

Without a deliberate architecture decision, Phase 4 risks spiralling into a heavyweight JavaScript SPA build — contradicting all of the above constraints.

---

## 2. Decision

**Adopt a Server-Side Rendering (SSR) first approach** with HTMX-driven partial page updates, minimal client JavaScript via Alpine.js, type-safe HTML generation via `a-h/templ`, Tailwind CSS v4 for styling, and WebSocket (Go stdlib) for real-time push.

This is not "no JavaScript" — it is the **right amount of JavaScript per interaction type**:

| Interaction type | Technology |
| --- | --- |
| Full page load (initial navigation) | `a-h/templ` server-rendered HTML |
| Partial page updates (form submit, list refresh) | HTMX 2 (`hx-post`, `hx-get`, `hx-swap`) |
| Component-local state (modal open, dropdown, countdown) | Alpine.js 3 (`x-data`, `x-show`, `x-bind`) |
| Real-time server push (balance update, income event) | WebSocket + HTMX Out-of-Band (OOB) swap |
| Layout, theme, responsiveness | Tailwind CSS v4 utility classes |

---

## 3. Stack Overview

### 3.1 `a-h/templ` — Type-Safe Server-Side Templates

**Version:** latest (`a-h/templ`)

```go
// Example: internal/ui/components/badge.templ
templ Badge(variant string, text string) {
    <span class={ "badge", "badge-" + variant }>{ text }</span>
}
```

**Rationale:**

- Templ compiles `.templ` files to typed Go functions. A typo in a component call is a **compile error**, not a runtime panic.
- Compared to `html/template`: templ provides autoescaping **and** compile-time type checking on template arguments. `html/template` is stringly-typed; a variable passed to the wrong slot fails at runtime.
- Compared to JSX/TSX: templ is pure Go — no Node.js, no Babel, no Webpack, no TypeScript compiler, no `node_modules`. A single `go generate` call regenerates all templates.
- Output is standard `io.Writer`-compatible; components compose naturally as Go functions.
- Design references: [a-h/templ documentation](https://templ.guide/).

### 3.2 HTMX 2 — Hypermedia-Driven Partial Updates

**Version:** 2.x (vendored, served via `embed.FS`)

```html
<!-- Submit form, replace the #invoice-list fragment without a full page reload -->
<form hx-post="/dashboard/transactions"
      hx-target="#transaction-list"
      hx-swap="outerHTML"
      hx-indicator="#spinner">
  ...
</form>
```

**Rationale:**

- HTMX extends HTML with `hx-*` attributes. Zero JavaScript is required in page source to trigger AJAX updates.
- No client-side routing, no state management library, no virtual DOM reconciliation.
- `hx-boost` upgrades anchor links and form submits to AJAX automatically, giving SPA-feel navigation with zero JS code.
- Out-of-Band (OOB) swaps (`hx-swap-oob`) allow a single server response to update multiple DOM fragments — critical for WebSocket push updates that must touch sidebar counters, list rows, and toast banners simultaneously.
- The `HX-Request: true` header sent on every HTMX request allows the server to return partial HTML fragments (no `<html>/<head>/<body>` wrapper) for HTMX requests, and full page responses for direct browser navigation.
- Compared to SPA frameworks: no JSON serialisation round-trip, no client-side routing to maintain, no bundle splitting needed.

### 3.3 Alpine.js 3 — Component-Local Reactivity

**Version:** 3.x (vendored, served via `embed.FS`)

```html
<!-- Modal with open/close state kept entirely in Alpine -->
<div x-data="{ open: false }">
  <button @click="open = true">Add Transaction</button>
  <div x-show="open" x-cloak @keydown.escape.window="open = false">
    <!-- modal content -->
  </div>
</div>
```

**Rationale:**

- Alpine (~15 KB minified+gzipped) fulfils the 5–10% of interactions that are inherently client-side: modals, dropdown menus, form validation UI, countdown timers (OTP resend), dark-mode toggle with `localStorage` persistence, and WebSocket reconnection backoff.
- It complements HTMX cleanly: HTMX handles server communication; Alpine handles local DOM state. Responsibilities do not overlap.
- Alpine is activated by HTML attributes, not JavaScript files — it is unobtrusive and degrades gracefully if disabled.
- Compared to Vue/React micro-frontends: a single `<script>` tag; no build step, no module bundler, no virtual DOM.

### 3.4 Tailwind CSS v4 — Utility-First Styling

**Version:** 4.x (`@tailwindcss/cli`)

```css
/* web/static/css/app.css */
@import "tailwindcss";

@theme {
  --color-brand-500: oklch(62.8% 0.2577 264.05);
  --font-sans: "Inter", system-ui, sans-serif;
  --radius-card: 0.75rem;
}
```

**Rationale:**

- Tailwind v4 is CSS-first: design tokens are defined directly in CSS via `@theme {}` blocks and become CSS custom properties automatically. No `tailwind.config.js` to maintain.
- Utility classes eliminate the semantic-naming problem of CSS modules ("what do I name this wrapper?") and prevent specificity battles.
- The `@tailwindcss/cli` binary generates a single optimised CSS file — only classes actually used in `.templ` files are emitted.
- Dark mode is supported via the `dark:` variant and a `<html data-theme="dark">` toggle controlled by Alpine + `localStorage`.
- Sub-10 KB final CSS for initial viewport is achievable with utility discipline.

### 3.5 WebSocket (stdlib `net/http`) — Real-Time Push

**Dependency:** `github.com/gorilla/websocket` (upgrader only — no framework)

```go
// internal/platform/ws/hub.go
type Hub struct {
    tenants map[string]map[*Client]struct{} // tenant_id → set of clients
    mu      sync.RWMutex
}

func (h *Hub) Broadcast(tenantID string, msg []byte) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    for c := range h.tenants[tenantID] {
        c.send <- msg
    }
}
```

**Rationale:**

- The API already pushes events (balance updates after transactions, income received after event). WebSocket allows these to propagate to the browser with sub-second latency without polling.
- Per-tenant rooms isolate broadcasts: a balance update for Tenant A never reaches Tenant B's browser. This is critical for data privacy.
- HTMX OOB swaps mean the WebSocket message **is HTML** — the server sends a rendered `<div id="account-balance" hx-swap-oob="true">` fragment directly. No client-side JSON parsing or DOM diffing required.
- Using Go stdlib `net/http` upgrader (via `gorilla/websocket`) keeps the library dependency surface minimal.
- The `Publisher` interface allows the hub backend to be swapped from in-process to Redis pub/sub in a future phase without client code changes.

---

## 4. Directory Layout

```
cmd/
    api/
        main.go                    # API binary entry point (existing)
    web/
        main.go                    # Web UI binary entry point (new)

internal/
    ui/
        layout/
            base.templ             # Root HTML shell: <html>, <head>, sidebar, topbar, content slot
            sidebar.templ          # Collapsible navigation sidebar
            topbar.templ           # Mobile header with hamburger, user avatar, theme toggle
            nav_item.templ         # Single navigation link with active-state highlighting
        components/
            button.templ           # Button (primary / secondary / danger / ghost variants)
            input.templ            # Text/number/email input with label + error message
            select.templ           # Native <select> + custom styled wrapper
            textarea.templ         # Multi-line text field
            checkbox.templ         # Accessible checkbox with label
            form_field.templ       # Wrapper: label + input + error (composable)
            modal.templ            # Alpine-driven overlay with backdrop
            drawer.templ           # Side-panel variant of modal (mobile-friendly)
            toast.templ            # Temporary notification (success / error / info)
            table.templ            # Responsive data table with sort headers
            card.templ             # Content card with optional header/footer slots
            badge.templ            # Status/category label (colour-coded)
            spinner.templ          # Loading indicator (CSS animation)
            skeleton.templ         # Content placeholder for loading states (HTMX request indicator)
            empty_state.templ      # Zero-data illustration + call-to-action
            currency_amount.templ  # Right-aligned monetary value formatted from int64 cents
            page_header.templ      # Page title + breadcrumb + primary action button row
        pages/
            auth/
                login.templ        # OTP request form
                verify.templ       # OTP code entry form
            dashboard/
                dashboard.templ    # Overview: net balance, recent transactions, income timeline
            accounts/
                list.templ
                detail.templ
            transactions/
                list.templ
                form.templ
            categories/
                list.templ
            master_purchase/
                list.templ
                form.templ
            investments/
                portfolio.templ
                position_detail.templ
            admin/
                users.templ
                tenants.templ
        handler/
            auth_handler.go        # GET /web/login, POST /web/auth/otp/request, POST /web/auth/otp/verify, POST /web/auth/logout
            dashboard_handler.go
            account_handler.go
            transaction_handler.go
            category_handler.go
            master_purchase_handler.go
            investment_handler.go
            admin_handler.go
        middleware/
            auth.go                # Cookie-based PASETO middleware for web routes
            ctx.go                 # Inject user + tenant structs into request context
        testutil/
            web_server.go          # newTestWebServer() for integration tests
            html.go                # HTML assertion helpers (AssertHasElement, AssertText)
        smoke_test.go              # Phase 4 baseline smoke tests (build tag: integration)

    platform/
        ws/
            hub.go                 # WebSocket broadcast hub (per-tenant rooms)
            client.go              # Per-connection read/write pumps with ping/pong
            publisher.go           # Publisher interface + InProcessPublisher implementation

web/
    embed.go                       # //go:embed static — embeds static/ into embed.FS
    static/
        css/
            app.css                # Tailwind v4 entrypoint: @import "tailwindcss" + @theme {}
        js/
            htmx.min.js            # Vendored HTMX 2 bundle
            alpine.min.js          # Vendored Alpine.js 3 bundle
            ws.js                  # WebSocket + HTMX OOB reconnect plugin (Alpine component)
        img/
            logo.svg
            favicon.ico
```

---

## 5. Binary Separation: `cmd/web/` vs `cmd/api/`

The web server runs as a **separate binary** from the API. Both are compiled from the same module and share `internal/` packages.

### Rationale

| Concern | API binary (`cmd/api/`) | Web binary (`cmd/web/`) |
| --- | --- | --- |
| Primary consumers | Mobile apps, third-party integrations, internal services | Browser users |
| Auth mechanism | `Authorization: Bearer <token>` header (PASETO) | `HttpOnly` cookie (`moolah_token`) |
| Response format | JSON | HTML (full page or HTMX fragment) |
| Rate limits | API-level throttling per tenant | Session-level throttling per cookie |
| Scale independently | High QPS API can scale without scaling the SSR renderer | SSR renderer can scale without scaling the API |
| Security surface | Different CSP, different CORS policy | Strict-origin CSP; no CORS needed |

### Shared internal packages used by `cmd/web/`

- `internal/service/` — all existing business logic services (account, transaction, category, investment, admin)
- `internal/domain/` — entity structs and interfaces
- `internal/platform/repository/` — SQLC-generated DB access
- `internal/platform/middleware/` — (reused where applicable; some middleware extended with cookie support)
- `pkg/paseto/` — token parsing (shared between header-bearer and cookie extraction)
- `pkg/ulid/` — ID generation
- `pkg/logger/` — structured logging

The web binary does **not** re-implement business logic. It calls the same service layer functions. There is no "proxy to the API" over HTTP — it shares Go function calls directly.

---

## 6. Authentication Strategy

### Token issuance

The existing `POST /auth/otp/verify` API endpoint continues to issue PASETO tokens. The web login flow:

1. User submits email at `POST /web/auth/otp/request` → calls `authService.SendOTP`.
2. User enters OTP code at `POST /web/auth/otp/verify` → calls `authService.VerifyOTP`, receives PASETO token.
3. Web handler sets the token as a **cookie** instead of returning it in a JSON body.

### Cookie specification

```
Set-Cookie: moolah_token=<paseto_v4_token>; HttpOnly; SameSite=Strict; Path=/; Secure; Max-Age=86400
```

| Attribute | Value | Reason |
| --- | --- | --- |
| `HttpOnly` | true | Inaccessible to JavaScript; mitigates XSS token theft |
| `SameSite=Strict` | true | Token is not sent on cross-site navigation; mitigates CSRF |
| `Secure` | true | HTTPS-only transmission |
| `Path=/` | / | Sent on all web requests |
| `Max-Age` | 86400 (24 h) | Matches the PASETO token TTL |

### Middleware

`internal/ui/middleware/auth.go` reads the cookie, parses the PASETO token using `pkg/paseto`, and injects `domain.User` + `domain.Tenant` into `context.Context`. Unauthenticated requests are redirected to `GET /web/login`.

For HTMX requests (`HX-Request: true`), an expired/missing token causes a `HX-Redirect` response header instead of a 302 redirect, so HTMX handles the navigation without breaking the partial update lifecycle.

---

## 7. Real-Time: WebSocket Architecture

### Hub design

```
┌─────────────────────────────────────────────────────┐
│  per-tenant rooms (map[tenantID]→set[*Client])       │
│                                                      │
│  Register(client)  ─────────────────────────────►   │
│  Unregister(client) ────────────────────────────►   │
│  Broadcast(tenantID, html_fragment) ─────────────►  │
└─────────────────────────────────────────────────────┘
         ▲                                  │
         │ HTTP Upgrade                     │ WebSocket write
   [Browser tab]                     [Browser tab]
```

### Event model

Events are **server-rendered HTML fragments** designed for HTMX OOB swap:

```go
// publisher.go: Broadcast sends pre-rendered HTML
type BalanceUpdatedEvent struct {
    AccountID string
    HTML      []byte // rendered by templ: <div id="account-balance-<id>" hx-swap-oob="true">...</div>
}
```

The browser Alpine plugin receives this message and injects it into the DOM:

```js
// web/static/js/ws.js
document.addEventListener('alpine:init', () => {
    Alpine.data('wsClient', (tenantID) => ({
        socket: null,
        retries: 0,
        connect() {
            this.socket = new WebSocket(`/ws?tenant=${tenantID}`)
            this.socket.onmessage = (e) => {
                // HTMX processes OOB swaps natively
                htmx.process(htmx.parseHTML(e.data))
            }
            this.socket.onclose = () => this.reconnect()
        },
        reconnect() {
            const delay = Math.min(1000 * 2 ** this.retries, 30000)
            this.retries++
            setTimeout(() => this.connect(), delay)
        },
    }))
})
```

### Event types

| Event | Trigger | OOB target selector | Payload |
| --- | --- | --- | --- |
| `balance_updated` | POST /transactions (any) | `#account-balance-{id}` | Rendered `CurrencyAmount` component |
| `income_received` | Income event marked `received` | `#income-timeline` | Rendered income row |
| `snapshot_ready` | Portfolio snapshot completed | `#portfolio-total` | Rendered `CurrencyAmount` + chart fragment |
| `toast` | Any server-side error or success | `#toast-container` | Rendered `Toast` component |

### Connection limits

`WS_MAX_CONNECTIONS_PER_TENANT` (default: `50`) limits connections collected per tenant to prevent a single tenant from consuming unbounded goroutines. Connections beyond the limit receive a `1008 Policy Violation` close frame.

### `Publisher` interface

```go
// internal/platform/ws/publisher.go
type Publisher interface {
    Publish(ctx context.Context, tenantID string, event Event) error
}

// InProcessPublisher — MVP implementation
type InProcessPublisher struct {
    hub *Hub
}

// RedisPubSubPublisher — future Phase 7 implementation (horizontal scaling)
```

---

## 8. Static Asset Strategy

### Development mode

Static files are served directly from the filesystem via `http.FileServer`. Templ's `--watch` mode and the Tailwind CLI `--watch` flag provide hot-reload of templates and styles respectively.

```sh
make web-dev  # runs templ generate --watch, tailwindcss --watch, and go run cmd/web/main.go with fsnotify live reload
```

### Production mode (default)

All static assets are embedded into the binary via `//go:embed static` in `web/embed.go`. This produces a **zero-external-file deployment**: a single binary contains the web server, all templates (compiled to Go functions), and all static assets (JS, CSS, images).

```go
// web/embed.go
package web

import "embed"

//go:embed static
var StaticFS embed.FS
```

The embedded file system is mounted at `/static/` in the web router:

```go
mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(web.StaticFS)))
```

### Content Security Policy

```
Content-Security-Policy:
  default-src 'self';
  script-src 'self' 'nonce-{per-request-nonce}';
  style-src 'self' 'nonce-{per-request-nonce}';
  connect-src 'self' wss://;
  img-src 'self' data:;
  frame-ancestors 'none'
```

- A per-request nonce is generated and injected into `<script>` and `<style>` tags by the `base.templ` layout.
- `wss://` is permitted in `connect-src` for the WebSocket connection.
- Inline scripts and styles are **not** permitted without a nonce.

### Cache-busting

Static filenames include a version hash (or are served with `ETag` + `Cache-Control: max-age=31536000, immutable`) to enable aggressive browser caching without stale-asset problems.

---

## 9. Build Pipeline

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  1. templ generate                                                           │
│     Reads *.templ files → generates *_templ.go companion files              │
│     (compile-error on type mismatches in template arguments)                 │
│                                                                              │
│  2. tailwindcss --input web/static/css/app.css --output web/static/css/dist.css │
│     Scans *.templ *.go for class usage → emits minimal CSS bundle           │
│                                                                              │
│  3. go build ./cmd/web/                                                      │
│     Compiles all Go + generated *_templ.go + embeds static/ via embed.FS    │
│     Output: bin/moolah-web                                                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Makefile targets (additions to existing Makefile)

| Target | Command(s) | Description |
| --- | --- | --- |
| `templ` | `templ generate` | Regenerate all `*_templ.go` files |
| `tailwind` | `tailwindcss --minify -i web/static/css/app.css -o web/static/css/dist.css` | Build optimised CSS |
| `web-build` | `templ` + `tailwind` + `go build ./cmd/web/` | Full production build of web binary |
| `web` | `go run ./cmd/web/` | Run web server in development |
| `web-dev` | watch mode for templ + tailwind + live-reload server | Development with hot-reload |

### CI pipeline changes

```yaml
# .github/workflows/ci.yml (addition)
- name: Install templ
  run: go install github.com/a-h/templ/cmd/templ@latest

- name: Install Tailwind CSS CLI
  run: curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 && chmod +x tailwindcss-linux-x64 && mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

- name: Generate templates
  run: make templ

- name: Build CSS
  run: make tailwind

- name: Build web binary
  run: go build ./cmd/web/

- name: Run web integration tests
  run: go test -tags integration -count=1 ./internal/ui/...
```

---

## 10. Testing Strategy

### Philosophy

No browser automation (Playwright / Selenium / Cypress) for Phase 4 MVP. A browser-testing harness takes 10× longer to write and maintain than `httptest` tests and flakes on CI. The `httptest`-based approach validates the HTTP contract and rendered HTML structure, with structural HTML assertions (not string matching) for robustness.

### Test harness

```go
// internal/ui/testutil/web_server.go
//go:build integration

type TestWebServer struct {
    URL   string   // e.g. "http://127.0.0.1:54321"
    WSURL string   // e.g. "ws://127.0.0.1:54321"
    DB    *sqlx.DB // direct access for test seeding
    t     testing.TB
}

func NewTestWebServer(t testing.TB) *TestWebServer
```

- Spins up a Postgres testcontainer (reusing `internal/testutil/containers`).
- Runs goose migrations.
- Builds the full dependency graph and starts an `httptest.NewServer`.
- Registers `t.Cleanup` to stop the server and terminate the container.

### HTML assertion helpers (no string matching)

```go
// internal/ui/testutil/html.go
func AssertHasElement(t testing.TB, body []byte, cssSelector string)
func AssertText(t testing.TB, body []byte, cssSelector, expected string)
func AssertAttr(t testing.TB, body []byte, cssSelector, attr, expected string)
```

Parses HTML with `golang.org/x/net/html` and walks the node tree. The helpers are selector-based (ID, element type, class), not position-based — they survive minor markup refactors.

### Smoke test matrix

| Test | Description |
| --- | --- |
| `TestWebSmoke_StaticAssets` | `htmx.min.js` and `alpine.min.js` served with correct `Content-Type` |
| `TestWebSmoke_AuthFlow` | OTP request → OTP verify → cookie set → dashboard loads → logout → redirect |
| `TestWebSmoke_ErrorPages_404` | Unknown route returns 404; `<h1>` contains "Page not found" |
| `TestWebSmoke_ErrorPages_HTMX_404` | Same route with `HX-Request: true` returns 404 + `HX-Retarget: #toast-container` |
| `TestWebSmoke_WebSocket_Rejected` | WebSocket upgrade without cookie → `403 Forbidden` |
| `TestWebSmoke_WebSocket_Accepted` | WebSocket upgrade with valid auth cookie → `101 Switching Protocols` |
| `TestWebSmoke_HTMXErrorToast` | HTMX error response contains toast fragment, not full page HTML |

All smoke tests carry `//go:build integration` and `t.Parallel()`.

---

## 11. Security Considerations

### XSS

- `a-h/templ` autoescapes all variable interpolations in HTML context by default. Explicit `templ.Raw()` is required to output unescaped HTML, and must only be used with content that has passed through a sanitiser (e.g. `bluemonday`).
- Content Security Policy (see §8) prevents inline script injection even if XSS were possible.

### CSRF

- `SameSite=Strict` on the session cookie prevents the browser from sending the cookie on cross-origin form submissions. This is effectively a modern CSRF defence without a separate token.
- HTMX forms are same-origin; WebSocket connections are verified using the session cookie.

### WebSocket Hijacking

- WebSocket upgrade is gated by the same `auth.go` middleware as HTTP routes. No valid session cookie, no upgrade.
- The origin check in the WebSocket upgrader is set to the app's domain only; connections from other origins are rejected with `403`.

### Cookie Security

- `HttpOnly` prevents JavaScript (including malicious scripts injected via third-party widgets) from reading the session token.
- `Secure` ensures the cookie is only transmitted over HTTPS.
- `Max-Age` bounds the session lifetime to 24 hours.

### Dependency supply chain

- HTMX and Alpine.js bundles are **vendored** into the repository (served via `embed.FS`). No CDN dependency. No risk of CDN-served script substitution attacks (`cdn.jsdelivr.net` compromise, etc.).

---

## 12. Consequences

### Positive

| Consequence | Detail |
| --- | --- |
| Single language | Frontend and backend are 100% Go. No context switching, no Node.js in the developer toolchain (only the Tailwind CSS binary is non-Go). |
| Type-safe templates | Template argument errors are caught at `go build` time, not in a production 500. |
| Minimal JS | ~15 KB Alpine + ~47 KB HTMX = ~62 KB total JS. Sub-100ms TTI on 3G. |
| Zero client-side routing | HTTP routes are the same for browser navigation and HTMX calls. SEO is free. |
| Real-time without polling | WebSocket push eliminates balance-staleness without wasting requests. |
| Embedded binary | `bin/moolah-web` is self-contained; deployment is `docker cp` + restart. |
| Testable with `httptest` | No browser / Selenium flake. Go tests are fast and deterministic. |

### Negative / Trade-offs

| Trade-off | Mitigation |
| --- | --- |
| `templ generate` must be run before `go build`; easy to forget | `Makefile` dependency chain; CI enforces it |
| Tailwind v4 requires a separate CLI binary | Pre-built binary downloaded in CI; no Node.js required |
| Full-page interactions trigger a server round-trip | HTMX partial swaps cover >95% of interactive surfaces; true full-reloads are rare |
| Complex client-side forms (multi-step wizards with local draft state) are awkward with pure HTMX | Alpine.js `x-data` handles draft state; server validates final submit |
| WebSocket reconnect logic lives in JavaScript |`ws.js` is small (~100 lines) and well-isolated |
| SEO for authenticated dashboard pages is irrelevant (private data) | SSR still delivers correct HTTP status codes for error pages |

---

## 13. Alternatives Considered

### 13.1 React / Next.js

**Rejected.**

| Criterion | Assessment |
| --- | --- |
| Language | TypeScript — a second language with its own toolchain, type system, and ecosystem. |
| Bundle size | A minimal Next.js app routes through >150 KB of framework JS before app code. |
| Build complexity | `webpack`/`turbopack`, `babel`, `eslint`, `prettier`, `package.json`, `node_modules`. |
| Team fit | Engineers are Go specialists. React+Next expertise would need to be hired or acquired. |
| Real-time | Requires a separate WebSocket state management layer (Zustand/Redux + socket.io). |
| Hydration | Hydration mismatch errors on SSR are a class of runtime bug with no Go equivalent. |

### 13.2 SvelteKit

**Rejected** for the same language-boundary and toolchain reasons as React/Next.js. Svelte is markedly better than React for bundle size, but still introduces a JavaScript compiler, a Node.js server, and a context switch from Go.

### 13.3 Pure `html/template` (no templ)

**Rejected.**

- Go's `html/template` is stringly-typed. Template arguments are `interface{}` / `any`; a typo or wrong type is a runtime panic, not a compile error.
- No component model — templates are text blocks without typed local state.
- Autoescaping is correct, but composability without `templ`'s component system requires fragile `{{template "name" .}}` call chains.
- `a-h/templ` compiles to `html/template` calls under the hood; it is a strict superset with no runtime overhead.

### 13.4 Full HTMX Without Alpine

**Rejected** as the sole client-side solution.

HTMX cannot efficiently manage:

- Modal open/close state (requires a server round-trip to toggle visibility, causing flicker).
- OTP countdown timer (purely client-side; no server involvement).
- Theme switching with `localStorage` persistence.
- WebSocket exponential backoff reconnect logic.

Alpine fills these ~10% of interactions that are inherently browser-local without adding meaningful bundle weight.

### 13.5 gRPC-Web / WebAssembly Frontend

**Rejected.**

These approaches require either a JavaScript gRPC-Web runtime or a WASM compilation pipeline — both add significant build complexity and are unsuitable for a household-scale application with a small engineering team.

---

## 14. Open Questions

| # | Question | Resolution |
| --- | --- | --- |
| 1 | Should the web server use the same port as the API or a separate port? | Separate port (default `:8081` for web, `:8080` for API). Same origin can be achieved via reverse proxy (Caddy/nginx) in production. |
| 2 | Is server-side rendering fast enough for dashboard with 100+ transactions? | Templ renders in microseconds; DB query time dominates. HTMX pagination and virtual scroll handles large lists client-side. |
| 3 | When to introduce Redis pub/sub for multi-pod WebSocket broadcasting? | Deferred to Phase 7 (horizontal scaling). `Publisher` interface makes the swap non-breaking. |
| 4 | Dark mode: OS preference or explicit toggle? | Both: Alpine reads `prefers-color-scheme` on first load; user can override via toggle (persisted in `localStorage`). |
| 5 | Internationalisation / i18n? | Not addressed in Phase 4. All UI strings are English. i18n is deferred post-MVP. |
