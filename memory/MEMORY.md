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

## Database
- Prepared statements in `pkg/db/db.go` `afterConn()` function
- VSDs types in `pkg/db/vsds.go`
- Errors: `db.ErrNotFound`, `db.ErrDuplicate`
