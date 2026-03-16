# Task 4.8 — Error Pages: 404, 403, 500

> **Roadmap Ref:** Phase 4 — UI Foundation & Design System › Error Handling
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-15
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Implement styled error pages for the three most important HTTP error states (404 Not Found, 403 Forbidden, 500 Internal Server Error) and wire them into the web server's central error handler. Error pages use the full base layout when the user is authenticated, or the auth layout when they are not. Each page has a clear, human-friendly message, a relevant illustration, and a call-to-action button to navigate back to safety.

---

## 2. Context & Motivation

Default Go `http.Error` responses return plain text. A finance app that serves an HTML UI must return consistently styled error pages that match the design system. Users who encounter errors should feel guided — not abandoned — with a clear path back to the dashboard or login page.

Error pages must also handle the HTMX case: when an error occurs during an HTMX partial request (identified by the `HX-Request: true` header), the server must return the error HTML with `HX-Retarget` and `HX-Reswap` headers to redirect the swap to the correct DOM element, or return a simple toast notification instead of a full page.

**Depends on:** Task 4.4 (base layout), Task 4.5 (Button, Card components).

---

## 3. Scope

### In scope

- [x] `internal/ui/pages/errors/not_found.templ` — 404 page.
- [x] `internal/ui/pages/errors/forbidden.templ` — 403 page.
- [x] `internal/ui/pages/errors/internal_error.templ` — 500 page.
- [x] `internal/ui/pages/errors/error_handler.go` — central `ErrorHandler` function and HTMX-aware response logic.
- [x] HTMX-aware error handling:
  - Full-page requests: render the full error page (with layout).
  - HTMX partial requests (`HX-Request: true`): return a toast-style error fragment with `HX-Retarget: #toast-container` and `HX-Reswap: beforeend`.
- [x] Register a custom 404 handler in `cmd/web/main.go` (Go's `ServeMux` default 404 replaced).
- [x] Wrap all web route handlers with a panic recovery middleware that renders the 500 page (instead of crashing the server).
- [x] Log all 500 errors with `slog.Error` including request path, method, and error details.

### Out of scope

- API error responses (the API returns JSON errors; that is unchanged).
- Custom error pages for other status codes (e.g. 429 rate limit, 503 — handled inline per handler).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                           | Purpose                                       |
| ------ | ---------------------------------------------- | --------------------------------------------- |
| CREATE | `internal/ui/pages/errors/not_found.templ`     | 404 page template                             |
| CREATE | `internal/ui/pages/errors/forbidden.templ`     | 403 page template                             |
| CREATE | `internal/ui/pages/errors/internal_error.templ`| 500 page template                             |
| CREATE | `internal/ui/pages/errors/error_handler.go`    | Central error response + HTMX detection       |
| CREATE | `internal/ui/pages/errors/error_handler_test.go` | Tests for full-page and HTMX error responses |
| MODIFY | `cmd/web/main.go`                              | Register custom 404; add recovery middleware  |

### Error page content design

| Page | Heading | Subtext | CTA |
| ---- | ------- | ------- | --- |
| 404  | "Page not found" | "The page you're looking for doesn't exist or has been moved." | "Go to Dashboard" |
| 403  | "Access denied" | "You don't have permission to view this page. If you think this is a mistake, contact your administrator." | "Go to Dashboard" |
| 500  | "Something went wrong" | "An unexpected error occurred. Our team has been notified. Please try again in a moment." | "Refresh page" + "Go to Dashboard" |

Each page also shows the numeric error code in a large, subtly styled number for quick visual identification.

### `ErrorHandler` function signature

```go
// internal/ui/pages/errors/error_handler.go

// RenderError writes the appropriate error response, respecting HTMX partial requests.
func RenderError(w http.ResponseWriter, r *http.Request, status int, props layout.BaseProps) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")

    isHTMX := r.Header.Get("HX-Request") == "true"

    if isHTMX {
        // Return a toast fragment instead of a full error page
        w.Header().Set("HX-Retarget", "#toast-container")
        w.Header().Set("HX-Reswap", "beforeend")
        w.WriteHeader(status)
        _ = components.Toast(toastPropsForStatus(status)).Render(r.Context(), w)
        return
    }

    w.WriteHeader(status)
    var page templ.Component
    switch status {
    case http.StatusNotFound:
        page = pages.NotFound(props)
    case http.StatusForbidden:
        page = pages.Forbidden(props)
    default:
        page = pages.InternalError(props)
    }
    _ = page.Render(r.Context(), w)
}
```

### Panic recovery middleware

```go
// internal/ui/middleware/recovery.go

func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                slog.ErrorContext(r.Context(), "panic recovered",
                    "path",   r.URL.Path,
                    "method", r.Method,
                    "panic",  rec,
                )
                errors.RenderError(w, r, http.StatusInternalServerError, basePropsFromCtx(r))
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

---

## 5. Acceptance Criteria

- [x] `GET /nonexistent-path` returns HTTP 404 with styled HTML error page.
- [x] `GET /dashboard` while unauthenticated redirects to `/login` (not a 403 error page — middleware handles this).
- [x] `GET /admin/tenants` as a non-sysadmin user returns HTTP 403 with styled HTML error page.
- [x] A handler that panics returns HTTP 500 with styled HTML error page (and does not crash the server).
- [x] All 500 panics are logged with `slog.Error` including path, method, and panic value.
- [x] HTMX partial requests to a 404 route receive a toast fragment + `HX-Retarget` header (not a full HTML page).
- [x] HTMX partial requests to a handler that returns 500 receive a toast fragment.
- [x] All error pages render correctly in both light and dark mode.
- [x] All three error pages are tested in `error_handler_test.go` for both full-page and HTMX modes.
- [x] `golangci-lint run ./internal/ui/pages/errors/...` passes.
- [x] `docs/ROADMAP.md` row 4.8 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change                                     |
| ---------- | ------ | ------------------------------------------ |
| 2026-03-15 | Agent  | Task implemented and verified (all tests pass) |
| 2026-03-15 | —      | Task created (new)                         |
