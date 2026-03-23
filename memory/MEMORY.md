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

## Database
- Prepared statements in `pkg/db/db.go` `afterConn()` function
- VSDs types in `pkg/db/vsds.go`
- Errors: `db.ErrNotFound`, `db.ErrDuplicate`
