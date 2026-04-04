# Phase 4 — REST API + `/static/` Handler + README

This implementation has been previously designed and approved.
Proceed with implementation directly.

Reference: `.claude/prompts/vsds-bundles-design.md`
Prerequisites: Phases 1–3 complete.

---

## Scope

- `pkg/db/bundles.go` — add API-side DB methods (create, list, get, delete, queue,
  update autoregen, update vsds config)
- `pkg/http/api/bundles/init.go` — route registration
- `pkg/http/api/bundles/list.go` — `GET /api/bundles`
- `pkg/http/api/bundles/get.go` — `GET /api/bundles/{id}`
- `pkg/http/api/bundles/create.go` — `PUT /api/bundles`
- `pkg/http/api/bundles/delete.go` — `DELETE /api/bundles/{id}`
- `pkg/http/api/bundles/generate.go` — `POST /api/bundles/{id}/generate`
- `pkg/http/api/bundles/patch.go` — `PATCH /api/bundles/{id}`
- `pkg/http/statichandler.go` — custom `/static/` handler
- `pkg/http/http.go` — register `/static/` when `Serve=true`; pass `BundlesConfig`
- `pkg/http/api/api.go` — register bundles subrouter
- `README.md` — nginx `gzip_static` note (if not already added in Phase 1)

---

## Additional DB methods — pkg/db/bundles.go

### `(p *DBPool) ListBundles() ([]Bundle, error)`

```sql
SELECT * FROM bundles.v_vsds_bundles ORDER BY id
```

Returns all bundles. Extend to union other type views when they are added.

### `(p *DBPool) GetBundle(id int) (*Bundle, error)`

```sql
SELECT * FROM bundles.v_vsds_bundles WHERE id = $1
```

Return `ErrNotFound` when 0 rows.

### `(p *DBPool) CreateVSDSBundle(...) (*Bundle, error)`

Call `SELECT * FROM bundles.create_vsds_bundle($1, $2, $3, $4, $5)` with
parameters: `name`, `autoregen`, `subtype`, `allprojects`, `projects int[]`.

Pass `projects` as a pgx `pgtype.Array[int32]` (or `[]int32`). When
`allprojects=true`, pass an empty/null array (the function ignores it).

Returns the `v_vsds_bundles` row for the new bundle (scan into `Bundle`).

### `(p *DBPool) DeleteBundle(id int) error`

```sql
DELETE FROM bundles.bundles WHERE id = $1
```

Return `ErrNotFound` when 0 rows affected.

### `(p *DBPool) QueueBundle(id int) error`

Sets `status='queued'` unless already `generating`:
```sql
UPDATE bundles.bundles
SET status = 'queued'
WHERE id = $1 AND status != 'generating'
```

Return `ErrAlreadyQueued` when 0 rows affected (meaning it was `generating`).
Also return `ErrNotFound` if the bundle doesn't exist — distinguish by first
checking existence or checking affected rows carefully.

### `(p *DBPool) UpdateBundleAutoregen(id int, autoregen bool) (*Bundle, error)`

```sql
UPDATE bundles.bundles SET autoregen = $2 WHERE id = $1
```

Return updated bundle via `GetBundle`. Return `ErrNotFound` when 0 rows.

### `(p *DBPool) UpdateVSDSBundle(...) (*Bundle, error)`

Parameters: `id int`, `name *string`, `autoregen *bool`, `subtype *string`,
`allprojects *bool`, `projects []int` (nil = unchanged).

This is a conditional update: only apply fields that are non-nil. Use a
transaction:
1. Fetch current vsds_bundles row
2. Apply non-nil fields via individual UPDATE statements or a dynamic approach
3. If any vsds-specific field changed (`subtype`, `allprojects`, `projects`):
   - DELETE from `vsds_bundle_projects` for this bundleid
   - Re-insert if new projects provided (and not allprojects)
   - `UPDATE bundles.bundles SET status='pending'` (file is stale)
4. Return updated row via `GetBundle`

---

## API response types

Define in the respective handler files (one response type per file, global
naming):

- `BundleListResponse = []db.Bundle`
- `BundleResponse = db.Bundle`

---

## pkg/http/api/bundles/init.go

```go
func Init(r *mux.Router) error {
    r.Handle("",
        w.NewAPIHandler().
            Get(listBundles).
            AuthPut(createBundle, w.IsAdmin),
    )
    r.Handle("/{id:[0-9]+}",
        w.NewAPIHandler().
            Get(getBundle).
            AuthDelete(deleteBundle, w.IsAdmin).
            AuthPatch(patchBundle, w.IsAdmin),  // check wrappers support PATCH
    )
    r.Handle("/{id:[0-9]+}/generate",
        w.NewAPIHandler().AuthPost(generateBundle, w.IsAdmin),
    )
    return nil
}
```

Check `pkg/http/wrappers/` to confirm PATCH is supported. If not, add it.

---

## Handler implementations

Each in its own file. All use `pkg/http/wrappers` per CLAUDE.md.

### list.go — GET /api/bundles (public)

Call `db.Pool.ListBundles()`. Return with `wrappers.Success`.

### get.go — GET /api/bundles/{id} (public)

Extract and validate `id` from `r.Vars`. Must be a positive integer.
Call `db.Pool.GetBundle(id)`. Map `ErrNotFound` → 404.

### create.go — PUT /api/bundles (admin)

Body:
```json
{
    "measurementtype": "vsds",
    "name": "...",
    "autoregen": false,
    "vsds": {
        "subtype": "surveypoints",
        "allprojects": false,
        "projects": [1, 2]
    }
}
```

Validate:
- `measurementtype` required; currently only `"vsds"` supported (return 400 for others)
- `name` required, non-empty
- VSDS: `subtype` required, must be `"surveypoints"` or `"surveys"`
- VSDS: `allprojects=false` requires `len(projects) > 0`
- VSDS: `allprojects=true` ignores `projects`

Call `db.Pool.CreateVSDSBundle(...)`. Return 201 with the new bundle.
Actually check if wrappers supports 201 or if 200 is used — follow existing pattern.

### delete.go — DELETE /api/bundles/{id} (admin)

Extract and validate id. Call `db.Pool.DeleteBundle(id)`.
Map `ErrNotFound` → 404. Return `wrappers.Success(nil)`.
File is NOT deleted (housekeeping is deferred).

### generate.go — POST /api/bundles/{id}/generate (admin)

Extract and validate id. Call `db.Pool.QueueBundle(id)`.
Map `ErrNotFound` → 404. Map `ErrAlreadyQueued` → 409.
On success: call `bundles.Signal()`, return `wrappers.Success(nil)`.

### patch.go — PATCH /api/bundles/{id} (admin)

Body (all fields optional):
```json
{
    "name": "...",
    "autoregen": true,
    "vsds": {
        "subtype": "surveys",
        "allprojects": true,
        "projects": [1, 2]
    }
}
```

Use pointer fields for all optional values so absent keys are distinguishable
from zero values.

Extract and validate id. Decode body. Call `db.Pool.UpdateVSDSBundle(...)`.
Map `ErrNotFound` → 404. Return updated bundle.

---

## pkg/http/statichandler.go

New file in package `http`. Serves files from `BundlesConfig.Path`.

```go
func newStaticBundleHandler(root string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Strip /static/ prefix
        // Reject path traversal (reject any ".." components)
        // Resolve to root + remaining path
        // If file ends in .json.gz: set headers
        //   Content-Type: application/json
        //   Content-Encoding: gzip
        // Serve the file with http.ServeContent or equivalent
    })
}
```

Security: reject any request path containing `..` — return 400. Only serve files
that exist in `root`; return 404 otherwise. No directory listing.

---

## pkg/http/http.go — modifications

1. Update `http.New` to accept `*config.Config` (or add `BundlesConfig` parameter)
   if not already done in Phase 1.

2. Register the static handler before the SPA catch-all when `Serve=true`:

```go
if cfg.Bundles.Serve {
    router.PathPrefix("/static/").
        Handler(newStaticBundleHandler(cfg.Bundles.Path))
}
```

---

## pkg/http/api/api.go

Add bundles subrouter after vsds:

```go
sr = r.PathPrefix("/bundles").Subrouter()
if err = apibundles.Init(sr); err != nil {
    return errors.Join(
        err, fmt.Errorf("Unable init bundles endpoints"))
}
```

---

## README.md

Ensure the nginx `gzip_static` note from Phase 1 is present. If Phase 1 was
implemented without it, add it now. No duplicate entries.

---

## Build verification

Run `gmake build`. Fix all compile errors before completing this phase.
