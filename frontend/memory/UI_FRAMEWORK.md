# UI Framework: PrimeNG

## Overview
The frontend uses **PrimeNG v20** with the **Aura** preset for all UI components. Bootstrap and ng-bootstrap were fully removed.

## Key Packages
- `primeng@^20` — component library
- `@primeuix/themes` — theme presets (Aura, Material, Lara, Nora) and `definePreset` utility
- `primeicons` — icon set used by PrimeNG components

## Theming & Dark Mode
- Theme is configured via `providePrimeNG()` in `src/main.ts` with the Aura preset
- Dark mode uses `darkModeSelector: 'system'` (the default), which generates `@media (prefers-color-scheme: dark)` CSS rules automatically
- No DOM manipulation needed — PrimeNG handles it via CSS media queries
- `ThemeService` (`src/app/auth/theme.service.ts`) only tracks the current theme string for display purposes (e.g. Settings page)

## CSS Variables
PrimeNG uses `--p-` prefixed CSS variables (design tokens):
- `--p-surface-ground` — page background
- `--p-text-color` — primary text
- `--p-text-muted-color` — secondary/muted text
- `--p-primary-color` — primary brand color
- `--p-content-border-color` — borders
- `--p-surface-100` — subtle backgrounds (sidebar)
- `--p-highlight-background` — hover/active highlight
- `--p-border-radius` — consistent border radius
- `--p-font-family` — font stack

## Component Usage Pattern
Each PrimeNG component is imported as a module in standalone components:
```typescript
import { CardModule } from 'primeng/card';
import { TagModule }  from 'primeng/tag';

@Component({
  imports: [CardModule, TagModule],
  ...
})
```

## Components Currently Used
- `ToolbarModule` (primeng/toolbar) — navbar
- `ButtonModule` (primeng/button) — buttons
- `MenuModule` (primeng/menu) — popup dropdown menus
- `CardModule` (primeng/card) — content cards
- `TagModule` (primeng/tag) — badges/labels
- `MessageModule` (primeng/message) — alerts/warnings
- `DividerModule` (primeng/divider) — horizontal dividers

## Layout
No grid library is used. Layouts use plain CSS flexbox. The `.sidebar` and `.page-wrapper` classes are defined in `src/styles.scss`.

## No PrimeFlex
PrimeFlex is deprecated. Layout is handled with custom CSS flexbox.

## PrimeNG Version Policy
Never use LTS-tagged PrimeNG versions (e.g. `20.5.0-lts`). Use the latest
non-LTS release (e.g. `^20.4.0`). LTS versions display a license banner
("You are using an LTS version of PrimeNG with an invalid license") because
the project does not have a PrimeStore license. Always pick the highest
non-LTS version tag for the current major.
