# Task 4.3 — Tailwind Configuration & Design Tokens

> **Roadmap Ref:** Phase 4 — UI Foundation & Design System › Styling
> **Status:** ✅ `done`
> **Last Updated:** 2026-03-15
> **Assignee:** —
> **Estimated Effort:** S

---

## 1. Summary

Define the Moolah visual language in `web/static/css/app.css` using Tailwind CSS v4 native CSS variables. Establish the full design token set — colour palette (light + dark mode), typography scale, spacing scale, border radii, shadow levels, and animation durations. The output is a single CSS file that all Templ components reference via Tailwind utility classes.

---

## 2. Context & Motivation

Tailwind CSS v4 uses a CSS-first configuration model: design tokens are declared as custom properties (`@theme`) inside the CSS file itself rather than a `tailwind.config.js`. This simplifies the build chain and keeps all design decisions in one place.

The UI must support:

- **Light and dark mode** — via `.dark` class on `<html>` (toggled by Alpine.js).
- **Mobile-first breakpoints** — `sm` (640px), `md` (768px), `lg` (1024px), `xl` (1280px).
- **Financial colour semantics** — green for income/credit, red for expense/debit, amber for pending, and a neutral blue-grey for UI chrome.
- **Accessibility** — WCAG AA contrast ratios for all text/background combinations.

**Depends on:** Task 4.2 (Tailwind CLI is installed and `app.css` stub exists).

---

## 3. Scope

### In scope

- [x] Define the `@theme` block in `web/static/css/app.css` with all design tokens.
- [x] Colour palette:
  - `brand-*`: primary blue-indigo scale (50–950) — used for primary actions, links, active states.
  - `neutral-*`: cool grey scale (50–950) — used for text, borders, backgrounds.
  - `success-*`: green scale — income, positive balance, confirmed states.
  - `danger-*`: red scale — expenses, negative balance, errors.
  - `warning-*`: amber scale — pending, overdue, caution states.
  - `surface-*`: light/dark surface layers (background, card, popover).
- [x] Semantic CSS variables for light/dark themes:
  - `--color-bg`, `--color-surface`, `--color-border`, `--color-text-primary`, `--color-text-muted`.
  - `--color-income`, `--color-expense`, `--color-pending`.
- [x] Typography scale: `font-sans` (Inter), `font-mono` (JetBrains Mono for amounts).
- [x] Spacing scale: default Tailwind 4 scale is sufficient; document deviations.
- [x] Border radius tokens: `radius-sm`, `radius-md`, `radius-lg`, `radius-xl`, `radius-full`.
- [x] Shadow tokens: `shadow-card`, `shadow-dropdown`, `shadow-modal`.
- [x] Animation/transition tokens: `duration-fast` (100ms), `duration-base` (200ms), `duration-slow` (350ms).
- [x] Google Fonts import or self-hosted font files for Inter + JetBrains Mono.
- [x] Dark mode toggle mechanism documented (Alpine.js reads `localStorage`, sets `.dark` on `<html>`).
- [x] Visual reference card `docs/design/tokens.md` listing all tokens and their values.

### Out of scope

- Component-specific styles (handled per-component in Tasks 4.4 and 4.5).
- Charts / data-visualisation colour palettes (handled in Phase 6 tasks).
- Print stylesheets.

---

## 4. Technical Design

### Files to create / modify

| Action | Path                           | Purpose                                          |
| ------ | ------------------------------ | ------------------------------------------------ |
| MODIFY | `web/static/css/app.css`       | Full design token definitions + base styles      |
| CREATE | `web/static/fonts/`            | Self-hosted Inter + JetBrains Mono woff2 files   |
| CREATE | `docs/design/tokens.md`        | Visual token reference (colours, spacing, etc.)  |

### `app.css` structure

```css
/* 1. Tailwind v4 base import */
@import "tailwindcss";

/* 2. Self-hosted fonts */
@font-face {
    font-family: "Inter";
    src: url("/static/fonts/inter-variable.woff2") format("woff2");
    font-weight: 100 900;
    font-display: swap;
}

@font-face {
    font-family: "JetBrains Mono";
    src: url("/static/fonts/jetbrains-mono-variable.woff2") format("woff2");
    font-weight: 100 900;
    font-display: swap;
}

/* 3. Design tokens via @theme */
@theme {
    /* Brand (blue-indigo) */
    --color-brand-50:  #eef2ff;
    --color-brand-500: #6366f1;
    --color-brand-600: #4f46e5;
    --color-brand-950: #1e1b4b;

    /* Success (green) */
    --color-success-100: #dcfce7;
    --color-success-500: #22c55e;
    --color-success-700: #15803d;

    /* Danger (red) */
    --color-danger-100: #fee2e2;
    --color-danger-500: #ef4444;
    --color-danger-700: #b91c1c;

    /* Warning (amber) */
    --color-warning-100: #fef3c7;
    --color-warning-500: #f59e0b;
    --color-warning-700: #b45309;

    /* Neutral */
    --color-neutral-50:  #f8fafc;
    --color-neutral-100: #f1f5f9;
    --color-neutral-200: #e2e8f0;
    --color-neutral-700: #334155;
    --color-neutral-800: #1e293b;
    --color-neutral-900: #0f172a;

    /* Typography */
    --font-sans: "Inter", ui-sans-serif, system-ui, sans-serif;
    --font-mono: "JetBrains Mono", ui-monospace, monospace;

    /* Radii */
    --radius-sm:   0.25rem;
    --radius-md:   0.375rem;
    --radius-lg:   0.5rem;
    --radius-xl:   0.75rem;
    --radius-full: 9999px;

    /* Shadows */
    --shadow-card:     0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
    --shadow-dropdown: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
    --shadow-modal:    0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);

    /* Transitions */
    --duration-fast: 100ms;
    --duration-base: 200ms;
    --duration-slow: 350ms;
}

/* 4. Semantic light-mode variables */
:root {
    --color-bg:           var(--color-neutral-50);
    --color-surface:      #ffffff;
    --color-border:       var(--color-neutral-200);
    --color-text-primary: var(--color-neutral-900);
    --color-text-muted:   var(--color-neutral-500);
    --color-income:       var(--color-success-700);
    --color-expense:      var(--color-danger-700);
    --color-pending:      var(--color-warning-700);
}

/* 5. Dark mode overrides */
.dark {
    --color-bg:           var(--color-neutral-900);
    --color-surface:      var(--color-neutral-800);
    --color-border:       var(--color-neutral-700);
    --color-text-primary: var(--color-neutral-50);
    --color-text-muted:   var(--color-neutral-400);
    --color-income:       var(--color-success-500);
    --color-expense:      var(--color-danger-500);
    --color-pending:      var(--color-warning-500);
}

/* 6. Base resets */
*, *::before, *::after { box-sizing: border-box; }
body { font-family: var(--font-sans); background-color: var(--color-bg); color: var(--color-text-primary); }
code, kbd, pre, samp { font-family: var(--font-mono); }
```

### Dark mode Alpine.js wiring (to document)

```javascript
// Injected in <html x-data x-bind:class="{ dark: $store.theme.dark }">
Alpine.store('theme', {
    dark: localStorage.getItem('theme') === 'dark' ||
          (!localStorage.getItem('theme') && window.matchMedia('(prefers-color-scheme: dark)').matches),
    toggle() {
        this.dark = !this.dark;
        localStorage.setItem('theme', this.dark ? 'dark' : 'light');
    }
});
```

---

## 5. Acceptance Criteria

- [ ] `make tailwind` compiles without errors and produces `web/static/css/app.min.css`.
- [x] All colour, typography, spacing, radius, shadow, and duration tokens are defined.
- [x] Light mode and dark mode semantic variables are correct (`--color-bg`, etc.).
- [x] Inter and JetBrains Mono fonts are served from `web/static/fonts/` (no CDN call).
- [ ] Dark mode toggling via `.dark` class is documented in `docs/design/tokens.md`.
- [ ] `docs/design/tokens.md` lists every token name, value in light mode, and value in dark mode.
- [x] Contrast ratios for primary text/bg combinations meet WCAG AA (verified manually).
- [ ] `docs/ROADMAP.md` row 4.3 updated to ✅ `done`.

---

## 6. Change Log

| Date       | Author | Change             |
| ---------- | ------ | ------------------ |
| 2026-03-15 | —      | Task created (new) |
