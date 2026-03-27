# Phase 5 — Frontend: Bundles Section

This implementation has been previously designed and approved.
Proceed with implementation directly.

Reference: `.claude/prompts/vsds-bundles-design.md`
Prerequisites: Phases 1–4 complete (all API endpoints available).

---

## Scope

New Angular components, service, and routes for the public `Bundles` top-level
navbar section. Admin functions are layered onto the same page.

Follows all existing patterns from `frontend/memory/MEMORY.md` and
`frontend/memory/CORRECTION_PATTERNS.md` — read both files before starting.

---

## New files

```
src/app/services/bundles.service.ts
src/app/components/bundles/
    bundles.component.ts
    bundles.component.html
    bundles.component.scss
    bundles-dashboard.component.ts
    bundles-dashboard.component.html
    bundles-dashboard.component.scss
    vsds-bundles/
        vsds-bundles.component.ts
        vsds-bundles.component.html
        vsds-bundles.component.scss
```

---

## Modified files

```
src/app/app.routes.ts                           add /bundles routes
src/app/components/navbar/navbar.component.ts   add Bundles section + breadcrumb
src/app/components/navbar/navbar.component.html add Bundles nav item
src/app/components/navbar/navbar.component.scss add any needed styles
```

---

## TypeScript interfaces

Define in `bundles.service.ts`:

```typescript
export interface Bundle {
  id: number;
  measurementtype: string;
  name: string;
  filename: string;
  generated_at: string | null;    // ISO8601
  autoregen: boolean;
  status: 'pending' | 'queued' | 'generating' | 'ready' | 'error';
  error_message: string | null;
  // vsds-specific
  subtype: string | null;
  allprojects: boolean | null;
  projects: number[] | null;
}

export interface CreateBundleRequest {
  measurementtype: string;
  name: string;
  autoregen: boolean;
  vsds?: {
    subtype: string;
    allprojects: boolean;
    projects: number[];
  };
}

export interface PatchBundleRequest {
  name?: string;
  autoregen?: boolean;
  vsds?: {
    subtype?: string;
    allprojects?: boolean;
    projects?: number[];
  };
}

export interface AppConfig {
  bundleBaseUrl: string;
}
```

---

## BundlesService — bundles.service.ts

Injectable service. Methods:

```typescript
getConfig(): Observable<AppConfig>
  // GET /api/config

listBundles(): Observable<Bundle[]>
  // GET /api/bundles

getBundle(id: number): Observable<Bundle>
  // GET /api/bundles/{id}

createBundle(body: CreateBundleRequest): Observable<Bundle>
  // PUT /api/bundles

deleteBundle(id: number): Observable<void>
  // DELETE /api/bundles/{id}

generateBundle(id: number): Observable<void>
  // POST /api/bundles/{id}/generate

patchBundle(id: number, body: PatchBundleRequest): Observable<Bundle>
  // PATCH /api/bundles/{id}
```

Cache `bundleBaseUrl` after the first `getConfig()` call. Expose it as a
synchronous getter once loaded (or return observable). Components use it to
construct download links as `bundleBaseUrl + bundle.filename`.

Follow the API service pattern from `api.service.ts` for response unwrapping
(`.data` field).

---

## Routing — app.routes.ts

`/bundles` is a public lazy-loaded section. No guard on the section shell.

```typescript
{
  path: 'bundles',
  loadComponent: () =>
    import('./components/bundles/bundles.component')
      .then(m => m.BundlesComponent),
  children: [
    {
      path: '',
      loadComponent: () =>
        import('./components/bundles/bundles-dashboard.component')
          .then(m => m.BundlesDashboardComponent)
    },
    {
      path: 'vsds',
      loadComponent: () =>
        import('./components/bundles/vsds-bundles/vsds-bundles.component')
          .then(m => m.VsdsBundlesComponent)
    }
  ]
}
```

---

## Section shell — bundles.component

Mirrors the VSDS shell pattern: sidebar with subsection links + `<router-outlet>`.

Sidebar subsections:
- **VSDS Bundles** — `/bundles/vsds` — always visible (public)

No guards on the shell or any subsection route — the section is public.
Admin-only UI elements are shown conditionally inside the component using
`authService.user?.isadmin`.

---

## Dashboard — bundles-dashboard.component

Landing page for `/bundles`. `p-carousel` of available subsections.
Subsections list is static (only VSDS for now). Follow the same pattern as
`vsds-dashboard.component`.

---

## VSDS Bundles subsection — vsds-bundles.component

This component is both the public download page and the admin management page.

### Public view (always shown)

A `p-table` listing all bundles returned by `listBundles()`. Columns:
- Name
- Subtype
- Projects ("All" when `allprojects=true`, comma-separated IDs otherwise)
- Status (badge: `p-tag` with severity based on status)
  - `pending` → secondary
  - `queued` / `generating` → info
  - `ready` → success
  - `error` → danger
- Last Generated (`generated_at` formatted as ISO8601 with space not T, seconds
  precision — e.g. `2026-03-27 14:22:05`)
- Download (link icon, `href = bundleBaseUrl + bundle.filename`; disabled/hidden
  when `status !== 'ready'`)

### Admin view (shown when `isAdmin`, layered on the same component)

Additional columns in the table (or action buttons in a dedicated column):
- Auto-regen toggle (`p-toggleButton` or checkbox; calls `patchBundle` on change)
- Generate button (`pi pi-play`, calls `generateBundle`; disabled when
  `status === 'generating'`)
- Delete button (`pi pi-trash`, danger; requires confirmation dialog before
  calling `deleteBundle`)

Confirmation dialogs on destructive actions (see `feedback_confirmation_dialogs.md`).

Add-bundle form (admin only, shown above or below table):
- Name (`p-inputText`, required)
- Measurement type: fixed to `vsds` for now (display as read-only label or hidden)
- Subtype (`p-dropdown`: `surveypoints` | `surveys`)
- All Projects checkbox (`p-checkbox`)
- Projects multi-select (`p-multiSelect`, listing `VSDSProject[]`; hidden when
  All Projects is checked)
- Auto-regen checkbox
- Submit calls `createBundle`, then refreshes the list

Use `VsdsService.listProjects()` (already exists) to populate the project
multi-select.

### Component lifecycle

On `ngOnInit`:
1. `bundlesService.getConfig()` to get `bundleBaseUrl` (if not already cached)
2. `bundlesService.listBundles()` to populate the table
3. (admin) `vsdsService.listProjects()` to populate project multi-select

Reload the bundle list after any create / delete / generate action.

---

## Navbar — add Bundles section

Follow the exact pattern documented in `frontend/memory/MEMORY.md` under
"Navbar Nav Section Pattern".

In `navbar.component.ts`:
- Add `bundlesMenuItems` property (NOT a getter — initialised in `ngOnInit`)
  with one item: `{ label: 'VSDS Bundles', routerLink: '/bundles/vsds' }`
- Extend `updateBreadcrumbs()` to handle `/bundles` and `/bundles/vsds`

In `navbar.component.html`:
- Add Bundles nav section (public — no `*ngIf` guard): section name link to
  `/bundles`, chevron dropdown using `bundlesMenuItems`
- Position: after VSDS, before any admin-only items (or at end — decide based on
  visual balance)

In `navbar.component.scss`:
- No special styles needed beyond existing `.nav-section` pattern unless required

---

## Status badge severity mapping

```typescript
statusSeverity(status: string): string {
  switch (status) {
    case 'ready':      return 'success';
    case 'error':      return 'danger';
    case 'generating':
    case 'queued':     return 'info';
    default:           return 'secondary';
  }
}
```

---

## Build verification

Run `gmake -C frontend/ build` first to catch TypeScript errors, then
`gmake build` for the full build. Fix all errors before completing this phase.
