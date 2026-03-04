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
- Shell: `components/vsds/vsds.component.ts` - sidebar + breadcrumb
- Dashboard: `components/vsds/vsds-dashboard.component.ts` - carousel landing page
- **Folders subsection** (`/vsds/folders`, protected: isAdmin):
  - `components/vsds/vsds-folders.component.ts`
  - Lists folders via GET /api/vsds/folders
  - Add folder: POST /api/vsds/folders with `{url: string}` (Google Drive URL)
  - Delete folder: DELETE /api/vsds/folders/{id}
  - Future: each folder row will expand to show document processing details
    - **TBD**: either collapsible row expansion in p-table OR route to /vsds/folders/{id}
    - When implemented, needs breadcrumb update in vsds.component.ts

## PrimeNG 20 Import Pattern
- Use `*Module` imports (e.g., `TableModule`, `CardModule`, `ButtonModule`)
- Or standalone classes (e.g., `Table`, `Card`, `Button`) - both work
- `CarouselModule` includes SharedModule (pTemplate works)
- `TableModule` includes SharedModule (pTemplate works)
- `DatePipe` from `@angular/common` can be imported standalone

## Breadcrumbs
- Shown only when a subsection is selected (NOT on dashboard/landing page)
- Placement: full-width bar ABOVE sidebar+content layout, directly below navbar
- Left-aligned by default — must NOT be inside the sidebar-offset content area
- Updates on NavigationEnd events via `hasSubsection` flag in shell component
- `p-breadcrumb` from `primeng/breadcrumb`
- home item: `{ icon: 'pi pi-home', routerLink: '/' }`
- Shell structure: `.vsds-shell` (margin-top:56px, flex-col) > `.breadcrumb-bar` + `.vsds-layout` (flex-row)

## Navbar Nav Section Pattern
- Section name (`<a class="nav-link section-name">`) + chevron (`<span class="section-chevron">`)
- Wrapped in `<span class="nav-section">` (display: flex)
- CSS in navbar.component.scss: `.nav-section`, `.section-name`, `.section-chevron`
- `vsdsMenuItems` is a getter for reactive permission filtering
