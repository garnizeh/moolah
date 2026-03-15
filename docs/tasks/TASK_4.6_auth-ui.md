# Task 4.6 — Authentication UI: OTP Request & Verify Pages

> **Roadmap Ref:** Phase 4 — UI Foundation & Design System › Authentication
> **Status:** 🔵 `backlog`
> **Last Updated:** 2026-03-15
> **Assignee:** —
> **Estimated Effort:** M

---

## 1. Summary

Implement the two authentication pages — OTP request (enter email) and OTP verify (enter 6-digit code) — as Templ templates wired to the existing API via HTMX form submissions. Set and read the PASETO token as an `HttpOnly` cookie so the web server can authenticate subsequent page requests without any JavaScript token management. Wire authentication middleware into the web server so that unauthenticated requests to protected pages redirect to the login page.

---

## 2. Context & Motivation

The API already provides `POST /auth/otp/request` and `POST /auth/otp/verify`. The web UI only needs to call these endpoints, receive the JWT, and store it as a cookie. All subsequent API calls from HTMX partials include the cookie automatically — no `Authorization` header management in JavaScript.

The authentication pages are the only pages that render without the full sidebar layout (they use a centred card layout instead).

**Auth flow:**

1. User visits `/login` → sees OTP request form.
2. Submits email → HTMX `POST /web/auth/otp/request` → web handler calls API → returns swap showing "check your email".
3. User enters OTP code → HTMX `POST /web/auth/otp/verify` → web handler calls API → on success: sets `HttpOnly` cookie + redirects to `/dashboard` via `HX-Redirect`.
4. On any authenticated page, web middleware reads cookie → validates JWT → injects user into context.
5. Logout: `POST /web/auth/logout` → clears cookie → redirects to `/login`.

**Depends on:** Task 4.2 (web server), Task 4.4 (layout — auth pages use a different centred shell), Task 4.5 (Button, Input, FormField, Spinner components).

---

## 3. Scope

### In scope

- [ ] `internal/ui/pages/auth/otp_request.templ` — email input form page.
- [ ] `internal/ui/pages/auth/otp_verify.templ` — 6-digit OTP input form page with countdown timer.
- [ ] `internal/ui/layout/auth_layout.templ` — centred card layout for unauthenticated pages (no sidebar).
- [ ] Web handler `internal/ui/pages/auth/auth_handler.go`:
  - `GET /web/login` — render OTP request page.
  - `POST /web/auth/otp/request` — proxy to API; return HTMX swap (success message or error).
  - `POST /web/auth/otp/verify` — proxy to API; set cookie on success; return `HX-Redirect: /dashboard`.
  - `POST /web/auth/logout` — clear cookie; return `HX-Redirect: /login`.
- [ ] Web auth middleware `internal/ui/middleware/auth.go`:
  - Reads `moolah_token` cookie; validates PASETO JWT.
  - Injects `*domain.User` and `*domain.Tenant` into `context.Context`.
  - Redirects unauthenticated requests to `/login` (returns `303 See Other`, not `401`).
- [ ] CSRF protection for all non-idempotent web form submissions (synchronizer token pattern or `SameSite=Strict` cookie policy — document choice).
- [ ] OTP verify page: Alpine.js countdown timer (10 minutes); "Resend code" link appears after 60s.
- [ ] Input auto-focus and auto-advance for the 6-digit OTP inputs (one `<input maxlength="1">` per digit or single 6-char input — document choice).
- [ ] Inline error display via HTMX `hx-swap="outerHTML"` on the form (replaces form with error-annotated version).
- [ ] Rate-limit error (429) displayed with friendly message: "Too many attempts. Please wait 15 minutes."
- [ ] Unit/integration tests for the auth handler.

### Out of scope

- Forgot/Reset flow (no passwords — OTP is the only auth method).
- Social login / OAuth (deferred).
- Two-factor / backup codes (deferred).
- Registration page (tenants are created by admin; users are invited by tenant owners — see Phase 5).

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                            | Purpose                                            |
| ------ | ----------------------------------------------- | -------------------------------------------------- |
| CREATE | `internal/ui/layout/auth_layout.templ`          | Centred card shell for unauthenticated pages       |
| CREATE | `internal/ui/pages/auth/otp_request.templ`      | Email entry page                                   |
| CREATE | `internal/ui/pages/auth/otp_verify.templ`       | OTP code entry page                                |
| CREATE | `internal/ui/pages/auth/auth_handler.go`        | Web auth HTTP handlers                             |
| CREATE | `internal/ui/pages/auth/auth_handler_test.go`   | Handler unit tests                                 |
| CREATE | `internal/ui/middleware/auth.go`                | Cookie-based auth middleware for web routes        |
| CREATE | `internal/ui/middleware/auth_test.go`           | Middleware tests                                   |
| MODIFY | `cmd/web/main.go`                               | Register auth routes + apply middleware            |

### Cookie specification

```
Name:     moolah_token
Value:    <PASETO v4 local token>
Path:     /
HttpOnly: true
Secure:   true  (false in development HTTP)
SameSite: Strict
MaxAge:   86400  (matches JWT TTL — 24h)
```

### Handler flow (OTP verify)

```go
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email string `form:"email"`
        Code  string `form:"code"`
    }
    if err := decodeForm(r, &req); err != nil {
        // render form with validation error via HTMX swap
        return
    }

    // Call API endpoint POST /auth/otp/verify
    apiResp, err := h.apiClient.VerifyOTP(r.Context(), req.Email, req.Code)
    if err != nil {
        // render form with error (rate-limit, invalid code, etc.)
        return
    }

    // Set HttpOnly cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "moolah_token",
        Value:    apiResp.Token,
        Path:     "/",
        HttpOnly: true,
        Secure:   h.cfg.IsProduction,
        SameSite: http.SameSiteStrictMode,
        MaxAge:   86400,
    })

    // HTMX redirect to dashboard
    w.Header().Set("HX-Redirect", "/dashboard")
    w.WriteHeader(http.StatusOK)
}
```

### OTP request page structure

```
auth_layout(title="Sign In")
  card (centred, max-w-md)
    Logo
    Heading: "Welcome to Moolah"
    Subheading: "Enter your email address to receive a sign-in code."
    form hx-post="/web/auth/otp/request" hx-swap="outerHTML"
      @form_field(label="Email address", input=input(type="email", name="email", required))
      @button(label="Send code", variant="primary", type="submit")
    small: "Don't have an account? Contact your administrator."
```

### OTP verify page structure

```
auth_layout(title="Enter your code")
  card (centred, max-w-md)
    Logo
    Heading: "Check your inbox"
    Subheading: "We sent a 6-digit code to {email}. Code expires in:"
    Alpine countdown timer (10 minutes)
    form hx-post="/web/auth/otp/verify" hx-swap="outerHTML"
      hidden input: email
      @form_field(label="Verification code", input=input(type="text", name="code",
                  maxlength="6", inputmode="numeric", autocomplete="one-time-code"))
      @button(label="Verify", variant="primary", type="submit")
    link (x-show="resendVisible"): "Resend code →" (hx-post to OTP request)
    link: "← Use a different email"
```

---

## 5. Acceptance Criteria

- [ ] `GET /web/login` renders the OTP request page (HTTP 200, HTML content).
- [ ] Submitting a valid email renders the "check your inbox" state via HTMX swap (no full-page reload).
- [ ] Submitting an invalid OTP code renders the form with an inline error message (no full-page reload).
- [ ] Submitting a valid OTP code sets the `moolah_token` `HttpOnly` cookie and triggers `HX-Redirect` to `/dashboard`.
- [ ] Auth middleware correctly rejects requests without a valid cookie (redirects to `/login`).
- [ ] Auth middleware injects user and tenant into context for valid tokens.
- [ ] Rate-limit error (429) shows the "Please wait 15 minutes" friendly message.
- [ ] OTP verify page countdown timer correctly counts down from 10 minutes using Alpine.js.
- [ ] "Resend code" link appears after 60 seconds.
- [ ] Logout endpoint clears the cookie and redirects to `/login`.
- [ ] `golangci-lint run ./internal/ui/...` passes.
- [ ] All handler and middleware tests pass.
- [ ] `docs/ROADMAP.md` row 4.6 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change             |
| ---------- | ------ | ------------------ |
| 2026-03-15 | —      | Task created (new) |
