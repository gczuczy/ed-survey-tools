# Frontend Notes

## Stack
- Angular 20, PrimeNG 20.4.x with Aura preset
- Standalone components (no modules)
- All components: separate .ts, .html, .scss files (no inline templates/styles)
- Block syntax: `@if`, `@for`, `@switch` (Angular 17+ built-in)

## Key Files
- Routing: `src/app/app.routes.ts`
- Navbar: `src/app/components/navbar/`
- Auth: `src/app/auth/` (auth.service.ts, auth.guard.ts, admin.guard.ts)
- API service: `src/app/services/api.service.ts`
- VSDS service: `src/app/services/vsds.service.ts`
- Global styles: `src/styles.scss` (`.sidebar`, `.nav-link`, `.page-wrapper`, `.text-muted`)

## Auth
- `AuthService`: `isLoggedIn`, `isConfigured`, `user` (UserInfo with `isadmin`, `isowner`)
- `authGuard`: requires login, triggers PKCE flow if not
- `adminGuard`: requires login+isAdmin (in admin.guard.ts)

## Sections Pattern (navbar)
- Public sections: always visible in navbar
- Each section: name link navigates to landing page + chevron dropdown for subsections
- Section shell component has sidebar + router-outlet + breadcrumb
- Sidebar shows subsection links filtered by permissions
- Landing page: dashboard with `p-carousel` of accessible subsections
- Subsection components loaded in router-outlet

## VSDS Section (vsds-basics branch)
- Path: `/vsds` (public section)
- Shell: `components/vsds/vsds.component.ts` - sidebar + router-outlet only (no breadcrumb)
- Dashboard: `components/vsds/vsds-dashboard.component.ts` - carousel landing page
- **Projects subsection** (`/vsds/projects`, public):
  - `components/vsds/vsds-projects.component.ts`
  - Split viewport: left = project list (p-listbox), right = zsamples panel
  - Admin left: add project form (PUT /api/vsds/projects, min 5 chars)
  - Right: zsamples shown as chips with X delete button (admin only)
  - Admin right: add single zsample (PUT .../zsamples/{zsample}), bulk replace via textarea (POST .../zsamples with []int)
  - All zsample mutations return updated VSDSProject; `updateProject()` syncs list + selection
  - `VSDSProject` interface: `{ id: number; name: string; zsamples: number[] }`
- **Folders subsection** (`/vsds/folders`, protected: isAdmin):
  - `components/vsds/vsds-folders.component.ts`
  - Lists folders via GET /api/vsds/folders
  - Add folder: POST /api/vsds/folders with `{url: string}` (Google Drive URL)
  - Delete folder: DELETE /api/vsds/folders/{id}
  - Process folder: POST /api/vsds/folders/{id}/process (empty body)
    - Process button: `pi pi-play`, success severity, disabled if `!canProcess(folder)`
    - `canProcess`: enabled if `!received_at` OR `finished_at != null`
    - On success: reloads folder list; on error: shows `processError` message
  - Future: each folder row will expand to show document processing details
    - **TBD**: either collapsible row expansion in p-table OR route to /vsds/folders/{id}
    - When implemented, needs breadcrumb update in navbar.component.ts (not vsds.component.ts)

## PrimeNG 20 Import Pattern
- Use `*Module` imports (e.g., `TableModule`, `CardModule`, `ButtonModule`)
- Or standalone classes (e.g., `Table`, `Card`, `Button`) - both work
- `CarouselModule` includes SharedModule (pTemplate works)
- `TableModule` includes SharedModule (pTemplate works)
- `DatePipe` from `@angular/common` can be imported standalone

## Breadcrumbs
- Shown only when a subsection is selected (NOT on dashboard/landing page)
- Placement: **inside navbar.component**, rendered right below the `<p-toolbar>` (per CLAUDE.md)
- Navbar host is `position: fixed` (wraps toolbar + breadcrumb bar); breadcrumb inherits the fixed context
- Logic in `navbar.component.ts`: subscribes to NavigationEnd, builds `breadcrumbItems` based on route
- `hasBreadcrumb` flag controls `@if` in navbar.component.html
- `p-breadcrumb` from `primeng/breadcrumb`; home item: `{ icon: 'pi pi-home', routerLink: '/' }`
- `ResizeObserver` on navbar host sets `--navbar-height` CSS var on `:root` for accurate page offset
- Pages/sections use `var(--navbar-height, 56px)` for margin-top (not hardcoded `56px`)
- When adding breadcrumbs for new sections, extend `updateBreadcrumbs()` in navbar.component.ts
- Section shell components (e.g. vsds.component) no longer contain breadcrumb logic

## Navbar Nav Section Pattern
- Section name (`<a class="nav-link section-name">`) + chevron (`<span class="section-chevron">`)
- Wrapped in `<span class="nav-section">` (display: flex)
- CSS in navbar.component.scss: `.nav-section`, `.section-name`, `.section-chevron`
- `vsdsMenuItems` is a property (NOT a getter) initialized in ngOnInit — a getter
  would return a new array every change detection cycle, causing p-menu to
  reinitialize and requiring double-click to navigate

## Recurring Correction Patterns
See `frontend/memory/CORRECTION_PATTERNS.md` for a full checklist of mistakes
that were historically corrected by the user after implementation.
