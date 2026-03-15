# Task 4.5 — Component Library

> **Roadmap Ref:** Phase 4 — UI Foundation & Design System › Components
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-15
> **Assignee:** GitHub Copilot
> **Estimated Effort:** L

---

## 1. Summary

Build the complete reusable Templ component library in `internal/ui/components/`. Every component is a typed Go function — no loose HTML snippets. The library covers: Button, Input, Select, Textarea, Checkbox, FormField (label + input + error), Modal, Drawer, Toast, Table (generic), Card, Badge, Spinner, Skeleton Loader, EmptyState, and CurrencyAmount (financial value with sign colouring). All components use only Tailwind utility classes and the design tokens defined in Task 4.3.

---

## 2. Context & Motivation

A well-designed component library is the single most important investment in a UI. Every page in Phases 5 and 6 will call these components directly. If a button style needs to change, it changes in one place. If a modal accessibility attribute is wrong, fixing it in one component fixes every modal in the app.

Templ's Go type system means components are statically type-checked at compile time — a missing required prop is a build error, not a runtime surprise.

**Depends on:** Task 4.3 (design tokens), Task 4.4 (layout context for modals/toasts).

---

## 3. Scope

### In scope

- [x] **Button** (`button.templ`) — variants: `primary`, `secondary`, `ghost`, `danger`; sizes: `sm`, `md`, `lg`; states: `loading` (shows spinner), `disabled`; supports `hx-*` attributes passthrough.
- [x] **Input** (`input.templ`) — text, email, number, date, search types; error state (red border + message); placeholder; required indicator.
- [x] **Select** (`select.templ`) — typed `[]SelectOption` slice; empty/placeholder option; error state.
- [x] **Textarea** (`textarea.templ`) — rows, maxlength, error state.
- [x] **Checkbox** (`checkbox.templ`) — label, description, checked state.
- [x] **FormField** (`form_field.templ`) — wraps any input with label, optional hint text, and error message. Used to compose standard form rows.
- [x] **Modal** (`modal.templ`) — Alpine.js-controlled; `x-show`; focus trap; `Escape` to close; configurable title, body slot, footer slot; sizes: `sm`, `md`, `lg`, `xl`.
- [x] **Drawer** (`drawer.templ`) — slides from the right; same Alpine.js pattern as modal; used for create/edit forms.
- [x] **Toast** (`toast.templ`) — four variants: `success`, `error`, `warning`, `info`; animated enter/leave; auto-dismiss; close button. Driven by `Alpine.store('toasts')`.
- [x] **Table** (`table.templ`) — `<table>` wrapper with styled header, row hover, and zebra-stripe option; column definitions via Go slice; empty state slot.
- [x] **Card** (`card.templ`) — surface card with optional header/footer slots; shadow token; padding variants.
- [x] **Badge** (`badge.templ`) — variants: `success`, `danger`, `warning`, `info`, `neutral`; sizes: `sm`, `md`.
- [x] **Spinner** (`spinner.templ`) — SVG animated ring; sizes: `sm`, `md`, `lg`; used inline in buttons and full-page loading.
- [x] **Skeleton** (`skeleton.templ`) — rectangular and circle variants for loading placeholders; pulse animation.
- [x] **EmptyState** (`empty_state.templ`) — icon + heading + description + optional CTA button; used when lists are empty.
- [x] **CurrencyAmount** (`currency_amount.templ`) — renders `int64` cents as formatted currency string; colours positive amounts green (`--color-income`) and negative red (`--color-expense`); accepts currency code.
- [x] **PageHeader** (`page_header.templ`) — page title, subtitle, and optional action slot (e.g. "Create" button); used at the top of every list page.
- [x] Component unit tests (`components_test.go`) — render each component and assert rendered HTML contains key structural elements and CSS classes.

### Out of scope

- Chart components (deferred to Phase 6).
- Data picker / calendar (deferred to Phase 5).
- File upload input (not needed in MVP).
- Rich text editor.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                                           | Purpose                            |
| ------ | ---------------------------------------------- | ---------------------------------- |
| CREATE | `internal/ui/components/button.templ`          | Button component                   |
| CREATE | `internal/ui/components/input.templ`           | Text/number/date input             |
| CREATE | `internal/ui/components/select.templ`          | Select dropdown                    |
| CREATE | `internal/ui/components/textarea.templ`        | Textarea                           |
| CREATE | `internal/ui/components/checkbox.templ`        | Checkbox                           |
| CREATE | `internal/ui/components/form_field.templ`      | Label + input + error wrapper      |
| CREATE | `internal/ui/components/modal.templ`           | Dialog overlay                     |
| CREATE | `internal/ui/components/drawer.templ`          | Side drawer                        |
| CREATE | `internal/ui/components/toast.templ`           | Notification toast                 |
| CREATE | `internal/ui/components/table.templ`           | Data table                         |
| CREATE | `internal/ui/components/card.templ`            | Content card                       |
| CREATE | `internal/ui/components/badge.templ`           | Status badge                       |
| CREATE | `internal/ui/components/spinner.templ`         | Animated loading spinner           |
| CREATE | `internal/ui/components/skeleton.templ`        | Content skeleton loader            |
| CREATE | `internal/ui/components/empty_state.templ`     | Empty list placeholder             |
| CREATE | `internal/ui/components/currency_amount.templ` | Financial value with sign colour   |
| CREATE | `internal/ui/components/page_header.templ`     | Page title + action slot           |
| CREATE | `internal/ui/components/components_test.go`    | Render tests for all components    |

### Component type examples

```go
// Button
type ButtonProps struct {
    Label    string
    Variant  string // "primary" | "secondary" | "ghost" | "danger"
    Size     string // "sm" | "md" | "lg"
    Type     string // "button" | "submit" | "reset"
    Loading  bool
    Disabled bool
    Class    string // extra Tailwind classes
    HTMX     map[string]string // hx-post, hx-target, etc.
}

// Badge
type BadgeProps struct {
    Label   string
    Variant string // "success" | "danger" | "warning" | "info" | "neutral"
    Size    string // "sm" | "md"
}

// CurrencyAmount
type CurrencyAmountProps struct {
    Cents        int64
    CurrencyCode string // "BRL", "USD", etc.
    ShowSign     bool   // always show + or - before the value
}

// Modal
type ModalProps struct {
    ID      string          // unique DOM ID, used by Alpine x-show
    Title   string
    Size    string          // "sm" | "md" | "lg" | "xl"
    Body    templ.Component
    Footer  templ.Component
}

// Table column definition
type TableColumn[T any] struct {
    Header string
    Cell   func(row T) templ.Component
}
```

### `CurrencyAmount` formatting logic

```go
// Format int64 cents → "R$ 1.234,56" (BRL) or "$ 1,234.56" (USD)
// Negative values → red class, positive → green class, zero → muted
func formatCents(cents int64, currency string) (display string, colorClass string) {
    // Use golang.org/x/text/currency + shopspring/decimal for formatting
    abs := cents
    if cents < 0 { abs = -cents }
    intPart := abs / 100
    decPart := abs % 100
    formatted := fmt.Sprintf("%d.%02d", intPart, decPart)
    // Apply locale formatting based on currency...
    switch {
    case cents > 0: colorClass = "text-[--color-income]"
    case cents < 0: colorClass = "text-[--color-expense]"
    default:        colorClass = "text-[--color-text-muted]"
    }
    return
}
```

### Accessibility requirements

| Component | Requirements |
| --------- | ------------ |
| Modal | `role="dialog"`, `aria-modal="true"`, `aria-labelledby` title id, focus trapped while open |
| Drawer | Same as Modal |
| Button (loading) | `aria-busy="true"`, `aria-disabled="true"` |
| Input (error) | `aria-invalid="true"`, `aria-describedby` error message id |
| Toast | `role="status"` (info/success) or `role="alert"` (error/warning) |
| Spinner | `role="status"`, `aria-label="Loading"` |
| Skeleton | `aria-hidden="true"` (decorative) |

---

## 5. Acceptance Criteria

- [x] All 17 component files compile via `make templ`.
- [x] `CurrencyAmount` formats BRL and USD correctly for positive, negative, and zero values.
- [x] `CurrencyAmount` applies correct colour class for each sign.
- [x] `Button` renders `aria-busy` and spinner icon when `Loading=true`.
- [x] `Modal` renders `role="dialog"` and `aria-modal="true"`.
- [x] `Input` renders `aria-invalid="true"` when `Error` is non-empty.
- [x] `components_test.go` has at minimum one_render test per component asserting key structural HTML.
- [x] All components use only Tailwind utility classes + design token CSS variables — no hardcoded colour hex codes.
- [x] `golangci-lint run ./internal/ui/components/...` passes.
- [x] `docs/ROADMAP.md` row 4.5 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author         | Change                                                |
| ---------- | -------------- | ----------------------------------------------------- |
| 2026-03-15 | —              | Task created (new)                                    |
| 2026-03-15 | GitHub Copilot | Task completed: 17 components implemented and tested. |
