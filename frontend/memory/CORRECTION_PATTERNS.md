# Frontend Correction Patterns

Compiled from analysis of all historical Claude session transcripts. These are
recurring mistakes that were corrected by the user, along with explicit rules
stated by the user. Use this as a checklist before finalizing any frontend
implementation.

---

## 1. Breadcrumbs - RECURRING (4+ sessions)

**Sessions:** 652b6318, 304556ad, 404f1792, 58421b18

### Mistakes made:
- Breadcrumb placed inside the section shell component (vsds.component), not
  in the navbar
- Breadcrumb not directly below the navbar (extra spacing between navbar and
  breadcrumb due to p-breadcrumb's internal padding: 1rem from Aura theme)
- Breadcrumb content showed static text ("extraction summary") instead of
  dynamic content (e.g. the folder's name)
- Breadcrumb shown on dashboard/landing page when no subsection is active
- Forgot to update breadcrumbs at all when adding a new view/page

### Rules:
- Breadcrumbs MUST live in `navbar.component` (below the `<p-toolbar>`)
- Breadcrumbs are NOT part of page/section components
- The p-breadcrumb Aura theme padding must be overridden to sit flush
- Breadcrumbs only appear when a subsection is active (not on landing)
- Dynamic breadcrumb label must use the actual data name (e.g. folder name),
  not a generic page title
- Extend `updateBreadcrumbs()` in `navbar.component.ts` for every new route

---

## 2. PrimeNG Dark Mode / Theme Compatibility - RECURRING (3+ sessions)

**Sessions:** c9ff3d41, ee4e04c8, 404f1792

### Mistakes made:
- Sidebar background used `--p-surface-100` which does not adapt in dark mode;
  correct token is `--p-content-background`
- Setting `color: var(--p-text-color)` in component-scoped CSS overrides the
  correctly-resolved white that a `p-card` establishes through CSS inheritance.
  In dark mode this causes black-on-black (or dark-on-dark) unreadable text
- Using `p-listbox` for project listing caused black-on-black text and green
  highlight on all items because `p-listbox` has hard-to-override Aura theme
  styles in dark mode
- `p-card` content area had white panel backgrounds in dark mode

### Rules:
- Avoid component-scoped `color: var(--p-text-color)` overrides — let
  elements inherit from the p-card or surface parent
- For list-style UI elements where PrimeNG theme fights color inheritance,
  prefer a plain `<ul>` or manually-styled div panels over `p-listbox`
- Always verify CSS variable tokens adapt in both light and dark mode before
  using them. Safe adaptive tokens: `--p-content-background`,
  `--p-text-color`, `--p-surface-ground`

---

## 3. Timestamp Display Format - RECURRING (2 sessions)

**Sessions:** f8e8581d, 404f1792

### Mistakes made:
- Used Angular `date:'short'` pipe which formats into locale format (not
  ISO8601)
- Displayed raw ISO string with the `T` separator intact

### Rules:
- All timestamps must be ISO8601, human-readable
- Transform: replace `T` with a space, seconds precision only
- Use `DatePipe` with format `'yyyy-MM-dd HH:mm:ss'` and `'UTC'` timezone
- Result example: `2026-03-05 19:37:00`
- Anything finer than seconds is not needed
- When displaying TZ-aware timestamps, make sure to include TZ indicator

---

## 4. p-button Disabled State and Severity Colors - RECURRING (2 sessions)

**Sessions:** f8e8581d, ee4e04c8

### Mistakes made:
- Using `[disabled]="true"` together with `[loading]="true"` — the disabled
  attribute overrides the severity color in PrimeNG Aura theme, causing the
  button to appear grey regardless of severity binding
- Using `pi-spin` class in `[icon]` binding inside a disabled button — the
  animation does not play

### Rules:
- PrimeNG `[loading]` already prevents clicks internally; do NOT add
  `[disabled]` on top of it when using `[loading]`
- For animated icons with color control, use `@if` blocks with three distinct
  button instances rather than one button with conditionally-bound severity
- State machine for process buttons:
  - Queued (received, not started): yellow clock icon (`pi-clock`,
    `severity="warning"`, no loading/disabled)
  - Running (started, not finished): animated blue spinner (`pi-spin
    pi-spinner`, `severity="info"`, no disabled so animation works)
  - Ready: green play button

---

## 5. Dropdown Menu Double-Click Issue - RECURRING (1 session + known fix)

**Session:** 732a00e8

### Mistake made:
- `vsdsMenuItems` (or any menu items array) was implemented as a getter,
  returning a new array on every change detection cycle. PrimeNG's `p-menu`
  sees the reference change and reinitializes, resetting focus. First click
  re-focuses the menu, second click actually fires the command.

### Rule:
- Menu item arrays for `p-menu` must be properties initialized in `ngOnInit`
  (or similar lifecycle hook), NOT getters
- Update the property only on actual state changes (e.g. NavigationEnd,
  login/logout events)

---

## 6. Sidemenu / Section Shell Layout - RECURRING (2 sessions)

**Sessions:** 304556ad, 652b6318, 58421b18

### Mistakes made:
- Sidemenu showed a heading with the section name ("VSDS") at the top; it
  should only list subsections, no heading
- The VSDS section shell was later found to not need a sidemenu at all;
  it was removed entirely

### Rules:
- Sidebar/sidemenu lists subsection links ONLY; no section-name heading
- When implementing a new section, check whether it actually needs a sidebar
  or just a router-outlet; do not add one by default
- If a section has a sidebar, its background must use an adaptive CSS token
  (e.g. `--p-content-background`) that works in both light and dark mode

---

## 7. p-table Row Expansion - RECURRING (4 user complaints in 1 session)

**Session:** 404f1792

### Mistakes made:
- `pRowToggler` directive applied to a `<p-button>` component does not
  reliably fire in PrimeNG 20 (binds to wrapper element, not inner button)
- `expandedRows` map management done incorrectly; first attempt: expand arrow
  turned but no content appeared; several iterations needed
- Wrong template name used for row expansion

### Rules:
- Do NOT use `pRowToggler` on `<p-button>` components in PrimeNG 20
- Manage row toggle manually: use a method that creates a new object
  reference on every toggle to guarantee change detection
- Verify PrimeNG version's exact API for `expandedRowKeys` before implementing
  row expansion

---

## 8. Layout: Icon Buttons Side-by-Side in Table Cells

**Session:** f8e8581d

### Mistake made:
- Play icon and trash icon were stacked vertically (above/below) instead of
  side by side, offsetting the table row height

### Rule:
- Action cell columns with multiple icon buttons need `display: flex;
  flex-direction: row` and sufficient `width` (e.g. `7rem` for 2 buttons)
- Always explicitly check table action column width when adding a second button

---

## 9. Dismissible Notifications

**Session:** 73c4460e

### Mistake made:
- Success notification ("Folder X added successfully") was rendered without a
  close/dismiss option

### Rule:
- Any success/error/info messages shown via `p-message` should be closable
- Use `[closable]="true"` and `(onClose)` to set the message variable to null

---

## 10. Angular Routing: SPA Base Href

**Session:** c9ff3d41

### Mistake made:
- `index.html` was missing `<base href="/">`. When the SPA handler served
  `index.html` for a deep URL (e.g. `/public-menu/option2`), the browser
  resolved relative JS/CSS paths relative to `/public-menu/` — all assets
  failed to load, resulting in a blank page on F5 reload

### Rule:
- `src/index.html` must always have `<base href="/">` — check this exists
  when setting up a new Angular project

---

## 11. Backend Go Struct JSON Tags - Causes Frontend Unreadability

**Session:** ee4e04c8

### Mistake made:
- Go struct for `VSDSProject` had no `json:""` struct tags. Go serialized
  field names as `ID`, `Name`, `ZSamples` (capitalized). TypeScript expected
  `id`, `name`, `zsamples`. Every `project.id` was `undefined`, causing
  `undefined === undefined` to make ALL items appear "selected" (green
  highlight) with no text rendered.

### Rule:
- ALWAYS add `json:"fieldname"` tags to Go structs that are serialized to
  the frontend
- When a frontend list renders blank or all items appear highlighted, check
  whether the backend struct has proper json tags

---

## 12. Newline Characters in Error Messages

**Session:** dd398e19

### Mistake made:
- Raw error messages containing `\n` newlines were displayed without any
  CSS treatment, so they appeared as a single long line

### Rule:
- Error message display elements must have `white-space: pre-wrap` in their
  CSS class to correctly render `\n` as visible line breaks

---

## 13. Lazy Loading of Routes

**Session:** 73c4460e

### Context:
- Initial implementation loaded all routes eagerly, causing the bundle to
  exceed 1.29 MB (budget was 1 MB)
- Lazy loading (`loadComponent`) reduced the initial bundle from 1.29 MB to
  738 KB

### Rule:
- All routes except the landing/home page should use `loadComponent` for
  lazy loading in `app.routes.ts`
- Do not add new routes as eagerly-loaded unless they are the root route

---

## 14. CLAUDE.md Must Be Read and Applied First

**Session:** 304556ad

### What happened:
- CLAUDE.md was loaded but its explicit UI rules (breadcrumb placement,
  sidemenu structure) were overlooked because the implementation was
  driven by pattern-matching from existing code rather than from the spec.

### Rule:
- Before implementing any frontend feature, explicitly cross-check against
  CLAUDE.md rules
- Do NOT delegate the initial codebase exploration entirely to an agent and
  then implement only from the agent's summary — the agent summary can lose
  important constraints
- Re-read CLAUDE.md after making a plan, before writing any code

---

## 15. Multiline String Handling in Error Lists

**Session:** dd398e19, e01438f2

### Context:
- The backend joins multiple error messages with `\n` separator when
  multiple system lookups fail for a single sheet/tab
- The frontend must render these correctly as separate lines

### Rule:
- Error message fields from the API may contain `\n` line separators
- Use `white-space: pre-wrap` on the containing CSS class

---

## 16. Confirmation Dialogs on Destructive Actions - RECURRING

**Session:** vsds-basics (2026-03-21, two separate corrections)

### Mistakes made:
- Delete variant button had no confirmation — user caught it
- Delete header check button had no confirmation — user caught it in the
  same session, immediately after the variant fix

### Rules:
- **Every** delete/remove button in the UI must show a confirmation dialog
  before executing. No exceptions, regardless of how "minor" the item.
- Use `ConfirmPopup` (anchored to the button) for inline row/item deletions:
  - Import `ConfirmPopupModule` from `primeng/confirmpopup`
  - Add `ConfirmationService` to component `providers`
  - Inject in constructor
  - Place `<p-confirmpopup />` once in the template (inside the relevant card)
  - Add a `confirmDeleteX(event: Event, id)` wrapper method that calls
    `confirmationService.confirm({ target: event.target, message, accept })`
  - Wire button to `(onClick)="confirmDeleteX($event, item.id)"` — never
    directly to the delete method
- Use `ConfirmDialog` (modal) for higher-stakes bulk or irreversible actions
  if needed, but `ConfirmPopup` is preferred for individual row deletions

---

## Summary Checklist for New Frontend Features

Before finishing any frontend implementation:

1. Are breadcrumbs handled in `navbar.component.ts`, not in the page component?
2. If a new route has a breadcrumb, does it show the actual data name (not a
   static title)?
3. Are timestamps formatted as `yyyy-MM-dd HH:mm:ss` (ISO8601, space not T)?
4. Do buttons with `[loading]` omit `[disabled]`?
5. Is any menu items array a property (not a getter)?
6. Does the action column in tables have `display: flex; flex-direction: row`?
7. Are success notifications closable (`[closable]="true"`)?
8. Are text elements using adaptive CSS tokens that work in dark mode?
9. Is any new list component (p-listbox, ul, etc.) readable in dark mode?
10. Do expanded rows in p-table use manual toggle (not pRowToggler on p-button)?
11. Are all new routes in app.routes.ts using lazy `loadComponent`?
12. Do all Go structs for API responses have `json:""` struct tags?
13. Do error message display elements have `white-space: pre-wrap`?
14. Does every delete/remove button go through a `ConfirmPopup` confirmation?
