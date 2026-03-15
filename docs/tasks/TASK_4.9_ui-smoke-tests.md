# Task 4.9 — UI Smoke / E2E Test Harness

> **Roadmap Ref:** Phase 4 — UI Foundation & Design System › Quality & Testing
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-15
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Establish the testing harness for web UI integration tests using Go's `httptest` package + HTML assertion helpers. Build a `newTestWebServer` helper (mirroring the API's `newTestServer` pattern) that spins up the full web stack against a Postgres testcontainer, and write the Phase 4 baseline smoke tests: toolchain smoke (static assets served), auth flow (OTP request → verify → cookie set → protected route accessible → logout), error page rendering (404, 403, 500), and WebSocket connection acceptance. All tests use the `integration` build tag.

---

## 2. Context & Motivation

The API smoke tests in `internal/server/smoke_test.go` validate the full API stack end-to-end. The web server needs the same treatment. Pure unit tests on Templ components cannot catch wiring bugs (wrong route path, missing middleware, cookie not set). The `httptest` harness gives us fast, hermetic, in-process integration tests that exercise the real HTTP stack without a browser.

HTML assertion helpers use Go's `golang.org/x/net/html` parser to check that critical DOM elements are present in rendered pages, without brittle string-matching.

**Depends on:** All prior Phase 4 tasks (4.2–4.8) — the harness tests everything built up to this point.

---

## 3. Scope

### In scope

- [ ] `internal/ui/testutil/web_server.go` — `newTestWebServer(t)` helper:
  - Starts a Postgres testcontainer (reuses `testutil/containers`).
  - Builds all dependencies (DB, auth middleware, hub, handlers).
  - Returns `*TestWebServer` with `.URL` and helper methods.
- [ ] `internal/ui/testutil/html.go` — HTML assertion helpers:
  - `AssertHasElement(t, body, selector)` — assert a CSS selector matches at least one element.
  - `AssertText(t, body, selector, expected)` — assert text content of matched element.
  - `AssertAttr(t, body, selector, attr, expected)` — assert attribute value on matched element.
  - `AssertHTTPStatus(t, resp, expected)` — assert HTTP response status code.
- [ ] `internal/server/smoke_test_web.go` (or `internal/ui/smoke_test.go`) — smoke tests:
  - `TestWebSmoke_StaticAssets` — `GET /static/js/htmx.min.js` returns 200 + `Content-Type: application/javascript`.
  - `TestWebSmoke_AuthFlow` — full OTP request → verify → cookie → dashboard → logout cycle.
  - `TestWebSmoke_ErrorPages` — 404 on unknown route, 403 on admin route as regular user, 500 via triggered panic handler.
  - `TestWebSmoke_WebSocketAccept` — WebSocket upgrade succeeds with valid auth cookie; upgrade fails without cookie.
  - `TestWebSmoke_HTMXErrorToast` — HTMX partial request to non-existent route returns toast fragment (not full HTML page).
- [ ] Cookie-based test helper: `withAuthCookie(req, token)` that sets the `moolah_token` cookie on a test request.
- [ ] Mock API client interface (`domain.APIClient` or equivalent) for unit tests that should not hit the real API.

### Out of scope

- Playwright / Selenium / Cypress browser-based E2E tests (deferred — overkill for MVP).
- Visual regression tests (screenshots).
- Performance/load tests.
- Testing every page in Phases 5 and 6 — each phase task owns its own UI tests.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                           | Purpose                                              |
| ------ | ---------------------------------------------- | ---------------------------------------------------- |
| CREATE | `internal/ui/testutil/web_server.go`           | `newTestWebServer` test helper                       |
| CREATE | `internal/ui/testutil/html.go`                 | HTML assertion utilities                             |
| CREATE | `internal/ui/smoke_test.go`                    | Phase 4 smoke tests (build tag: integration)         |

### `TestWebServer` type

```go
// internal/ui/testutil/web_server.go
//go:build integration

package uitestutil

type TestWebServer struct {
    URL    string
    DB     *sqlx.DB       // direct DB access for seeding
    Client *http.Client   // pre-configured with cookie jar
    t      testing.TB
}

func NewTestWebServer(t testing.TB) *TestWebServer {
    t.Helper()
    // 1. Start Postgres testcontainer (or reuse shared one)
    // 2. Run migrations
    // 3. Build full dependency graph (repos, services, handlers, hub)
    // 4. Start httptest.NewServer(mux)
    // 5. Register t.Cleanup to stop server
    return &TestWebServer{ ... }
}

// WithAuthCookie seeds a user+tenant, performs OTP verify, and returns a client
// that has the moolah_token cookie set.
func (s *TestWebServer) WithAuthCookie(t testing.TB) *http.Client { ... }
```

### HTML assertion helpers

```go
// internal/ui/testutil/html.go

import "golang.org/x/net/html"

// AssertHasElement fails the test if no element matching cssSelector is found in body.
func AssertHasElement(t testing.TB, body []byte, cssSelector string) {
    t.Helper()
    doc, err := html.Parse(bytes.NewReader(body))
    require.NoError(t, err)
    node := querySelector(doc, cssSelector) // custom CSS selector walker
    require.NotNil(t, node, "expected element matching %q to exist", cssSelector)
}

// AssertText fails if the text content of the matched element != expected.
func AssertText(t testing.TB, body []byte, cssSelector, expected string) {
    t.Helper()
    // ...
}
```

### Smoke test structure

```go
//go:build integration

package uismoke_test

func TestWebSmoke_AuthFlow(t *testing.T) {
    t.Parallel()
    srv := uitestutil.NewTestWebServer(t)

    // 1. OTP request
    resp := httpPost(t, srv, "/web/auth/otp/request", map[string]string{"email": testEmail})
    AssertHTTPStatus(t, resp, http.StatusOK)

    // 2. Get OTP from DB (test seeding helper)
    code := seeds.GetLatestOTPCode(t, srv.DB, testEmail)

    // 3. OTP verify
    resp = httpPost(t, srv, "/web/auth/otp/verify", map[string]string{
        "email": testEmail, "code": code,
    })
    AssertHTTPStatus(t, resp, http.StatusOK)
    require.NotEmpty(t, cookieValue(resp, "moolah_token"))

    // 4. Authenticated request to dashboard
    client := withCookieJar(resp)
    resp = httpGet(t, client, srv.URL+"/dashboard")
    AssertHTTPStatus(t, resp, http.StatusOK)
    body := readBody(t, resp)
    AssertHasElement(t, body, "nav[id='sidebar']")

    // 5. Logout
    resp = httpPost(t, client, srv.URL+"/web/auth/logout", nil)
    AssertHTTPStatus(t, resp, http.StatusOK)
    require.Empty(t, cookieValue(resp, "moolah_token"))

    // 6. Protected route now redirects
    resp = httpGet(t, http.DefaultClient, srv.URL+"/dashboard")
    AssertHTTPStatus(t, resp, http.StatusSeeOther)
}

func TestWebSmoke_ErrorPages(t *testing.T) {
    t.Parallel()
    srv := uitestutil.NewTestWebServer(t)

    // 404
    resp := httpGet(t, http.DefaultClient, srv.URL+"/this-does-not-exist")
    AssertHTTPStatus(t, resp, http.StatusNotFound)
    body := readBody(t, resp)
    AssertHasElement(t, body, "h1")
    AssertText(t, body, "h1", "Page not found")

    // HTMX 404 → toast fragment
    req, _ := http.NewRequest(http.MethodGet, srv.URL+"/this-does-not-exist", nil)
    req.Header.Set("HX-Request", "true")
    resp, _ = http.DefaultClient.Do(req)
    AssertHTTPStatus(t, resp, http.StatusNotFound)
    require.Equal(t, "#toast-container", resp.Header.Get("HX-Retarget"))
}

func TestWebSmoke_WebSocketAccept(t *testing.T) {
    t.Parallel()
    srv := uitestutil.NewTestWebServer(t)

    // Without auth cookie → rejected
    _, resp, err := websocket.DefaultDialer.Dial(srv.WSURL+"/ws", nil)
    require.Error(t, err)
    require.Equal(t, http.StatusForbidden, resp.StatusCode)

    // With valid auth cookie → accepted
    authClient := srv.WithAuthCookie(t)
    header := http.Header{"Cookie": []string{authCookieHeader(authClient, srv.URL)}}
    conn, _, err := websocket.DefaultDialer.Dial(srv.WSURL+"/ws", header)
    require.NoError(t, err)
    require.NoError(t, conn.Close())
}
```

---

## 5. Acceptance Criteria

- [ ] `newTestWebServer` starts successfully against a Postgres testcontainer.
- [ ] `TestWebSmoke_StaticAssets` passes: `htmx.min.js` and `alpine.min.js` are served with correct `Content-Type`.
- [ ] `TestWebSmoke_AuthFlow` passes: full OTP → cookie → dashboard → logout cycle.
- [ ] `TestWebSmoke_ErrorPages` passes (404 full-page, 404 HTMX toast, 403, 500).
- [ ] `TestWebSmoke_WebSocketAccept` passes: unauthenticated upgrade rejected; authenticated upgrade accepted.
- [ ] `TestWebSmoke_HTMXErrorToast` passes: HTMX partial request error returns `HX-Retarget` header.
- [ ] All HTML assertions use the structural parser (not string matching).
- [ ] All tests have `t.Parallel()`.
- [ ] `go test -tags integration ./internal/ui/...` passes in CI.
- [ ] `docs/ROADMAP.md` row 4.9 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change             |
| ---------- | ------ | ------------------ |
| 2026-03-15 | —      | Task created (new) |
