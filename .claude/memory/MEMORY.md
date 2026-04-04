# Backend Notes

## Project Structure
- Backend: Go, root directory
- Frontend: Angular 20, `frontend/` directory
- Build: `gmake build` (full), `gmake -C frontend/ build` (frontend only)
- Database: PostgreSQL, schemas in `sql/`, pkg in `pkg/db/`

## API Patterns
- All responses: `{"status": "success|error", "message": "...", "data": {}}`
- Response types globally defined in `pkg/http/api/<section>/`
- Wrappers: `wrappers.Success(data)`, `wrappers.NewError(err, httpCode)`
- Route registration via `w.NewAPIHandler().AuthGet(..., w.IsAdmin)` etc.
- URL vars auto-decoded, accessed via `r.Vars["key"]`

## VSDS API Endpoints
- `GET /api/vsds/folders` - list folders (AuthGet, IsAdmin)
- `POST /api/vsds/folders` - add folder by URL (AuthPost, IsAdmin)
- `DELETE /api/vsds/folders/{id}` - delete folder (AuthDelete, IsAdmin)
- `POST /api/vsds/folders/{id}/process` - queue folder processing (AuthPost, IsAdmin); returns 409 if already queued/in-progress
- `GET /api/vsds/projects`, `PUT /api/vsds/projects` etc.

### Variant Validation Endpoint (admin-only, backend complete as of 2026-03-23)
- `POST /api/vsds/projects/{id}/variants/validate` — validates a variant
  definition against a live Google Sheet URL; no DB writes
- Request body: `{ url: string, variant: { name, header_row,
  sysname_column, zsample_column, syscount_column, maxdistance_column,
  checks: [{col, row, value}] } }`
- Response: `{ tabs: [{ name, rows[][], checks[{col,row,expected,actual,ok}],
  matched }] }` — rows capped at 50×15
- Handler: `pkg/http/api/vsds/validate.go` `validateVariant`
- Helpers added: `gcp.ExtractSheetID(url)`, `gcp.Sheet.Cols()`,
  `pvsds.EvalChecks(checks, sheet)`, `pvsds.CheckInput`, `pvsds.CheckResult`

### Sheet Variant CRUD (admin-only, backend complete as of 2026-03-21)
- `GET /api/vsds/projects/{id}/variants` - list variants for project
- `PUT /api/vsds/projects/{id}/variants` - add variant
- `POST /api/vsds/projects/{id}/variants/{vid}` - update variant
- `DELETE /api/vsds/projects/{id}/variants/{vid}` - delete variant
- `PUT /api/vsds/projects/{id}/variants/{vid}/checks` - add check
- `DELETE /api/vsds/projects/{id}/variants/{vid}/checks/{cid}` - delete check
- All handlers in `pkg/http/api/vsds/variants.go`
- DB methods: `ListVariants`, `AddVariant`, `UpdateVariant`, `DeleteVariant`, `AddVariantCheck`, `DeleteVariantCheck` in `pkg/db/vsds.go`
- View: `vsds.v_spreadsheetvariants` in `sql/vsds_views.sql`
- Coordinates: 0-indexed (no transform in backend)

## VSDS Rho Display Convention
- Raw DB value `rho` is in **systems/ly³** (formula: `corrected_n / (4/3·π·maxdistance³)`,
  `maxdistance` ≤ 20 ly — see `sql/vsds_views.sql`).
- **Default display unit: systems/(10 Ly)³** — multiply raw value by 1000
  (1 box = 10×10×10 ly = 1000 ly³). Example: 0.001283 → 1.283 systems/box.
- Any UI component showing rho **must** include a toggle to switch to raw systems/ly³
  (unnormalized).

## VSDS Coordinate System
- Galactic coordinates are Sol-centered (0,0,0 = Sol).
- Galactic center is at `(25.21875, -20.90625, 25899.96875)` in Sol-centered
  x,y,z.
- Both `vsds.v_surveypoints` and `vsds.v_surveys` expose GC-relative coords:
  - `gc_x = x - 25.21875`
  - `gc_y = y + 20.90625`  (individual points only)
  - `gc_z = z - 25899.96875`
- `vsds.v_surveys` centroid is 2D (galactic plane): only `x`/`z` and `gc_x`/
  `gc_z` — `y` (galactic height) is omitted from the centroid.

## Database
- Prepared statements in `pkg/db/db.go` `afterConn()` function
- VSDs types in `pkg/db/vsds.go`
- Errors: `db.ErrNotFound`, `db.ErrDuplicate`

## SPA Handler MIME Types
`pkg/http/spahandler.go` has an `init()` that calls `mime.AddExtensionType`
for every extension used by the embedded frontend bundle.
**Why:** `tarfs.file` does not implement `io.Seeker`. When `net/http` doesn't
know the Content-Type (e.g. `.woff2` missing from minimal Linux MIME DB), it
falls back to content sniffing: reads 512 bytes then seeks back — that seek
fails, returning HTTP 500. CSS/JS worked fine; font files did not.
**Rule:** If new file types are added to the Angular build output, add the
corresponding `mime.AddExtensionType` entry to `init()`. Current set:
`.css`, `.eot`, `.html`, `.ico`, `.js`, `.svg`, `.ttf`, `.woff`, `.woff2`.

## GDPR / Personal Data
App is EU-based and EU-served — GDPR applies.
**Personal data stored:**
- `common.cmdrs.name` — CMDR in-game name, from FDev CAPI (login) and from
  processed survey spreadsheets (passive contributor, may not have logged in)
- `common.cmdrs.customerid` — FDev numeric account ID, only for logged-in users

**Legal basis:**
- Logged-in users: contract performance (Art. 6(1)(b))
- Spreadsheet-sourced CMDR names: legitimate interest (Art. 6(1)(f))

**Contact for data subject rights:** GitHub (open an issue / contact via profile)

**Erasure (self-delete in user Settings):** nullifies `common.cmdrs.customerid`
(sets to NULL); keeps CMDR name for survey attribution (legitimate interest
exemption); must also invalidate the user's Redis session.
**INVARIANT**: any future personal data field added to the schema MUST be
included in this removal before merging.

**Already implemented (2026-03-25):**
- Session cookie flags: `HttpOnly`, `SameSite=Strict`, `Secure`
  (config-driven; `sessions.secure: true` in prod, off by default for dev)
- Removed full FDev userinfo from INFO log in `pkg/http/api/auth/callback.go`

**Still to implement:**
- Cookie notice banner (one-time, localStorage-dismissed)
- Privacy notice (footer link → p-dialog)
- Self-delete endpoint (backend) + Settings UI (frontend)
