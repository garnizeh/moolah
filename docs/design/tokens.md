# Design Tokens Reference

This document serves as the visual reference for the Moolah design system tokens defined in `web/static/css/app.css`.

## 1. Typography

| Token | Family | Font Stack |
|-------|--------|------------|
| `--font-sans` | Inter | "Inter", ui-sans-serif, system-ui, sans-serif |
| `--font-mono` | JetBrains Mono | "JetBrains Mono", ui-monospace, monospace |

## 2. Colors (Semantic)

These variables change based on the user's theme preference (Light/Dark).

| Variable | Light Value | Dark Value | Purpose |
|----------|-------------|------------|---------|
| `--color-bg` | `neutral-50` | `neutral-950` | Main application background |
| `--color-surface` | `#ffffff` | `neutral-900` | Card, modal, and popover surfaces |
| `--color-border` | `neutral-200` | `neutral-800` | UI borders and dividers |
| `--color-text-primary` | `neutral-900` | `neutral-50` | Primary text |
| `--color-text-muted` | `neutral-500` | `neutral-400` | Secondary/de-emphasized text |
| `--color-income` | `success-700` | `success-500` | Positive balances, income transactions |
| `--color-expense` | `danger-700` | `danger-500` | Negative balances, expense transactions |
| `--color-pending` | `warning-700` | `warning-500` | Pending or caution states |

## 3. Dark Mode Toggle

Dark mode is activated by adding the `.dark` class to the `<html>` element. This is managed via Alpine.js:

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

## 4. Spacing & Shape

### Border Radius

- `radius-sm`: `0.25rem` (4px)
- `radius-md`: `0.375rem` (6px)
- `radius-lg`: `0.5rem` (8px)
- `radius-xl`: `0.75rem` (12px)
- `radius-full`: `9999px`

### Shadows

- `shadow-card`: Suble lift for dashboard cards.
- `shadow-dropdown`: Medium lift for menus.
- `shadow-modal`: High lift for dialogs.

## 5. Motion

- `duration-fast`: `100ms` (hover states, toggle switches)
- `duration-base`: `200ms` (fade-ins, scale transitions)
- `duration-slow`: `350ms` (slide-overs, page transitions)
