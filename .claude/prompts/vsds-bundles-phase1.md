# Phase 1 — DB Schema + Config + `/api/config`

This implementation has been previously designed and approved.
Proceed with implementation directly.

Reference: `.claude/prompts/vsds-bundles-design.md`

---

## Scope

- `sql/bundles_struct.sql` — new `bundles` schema + 3 tables
- `sql/bundles_views.sql` — `v_bundles`, `v_vsds_bundles`
- `sql/bundles_funcs.sql` — `create_vsds_bundle()`, `queue_autoregen_bundles()`
- `sql/_all.sql` — import bundles files after vsds (bundles has FK into vsds.projects)
- `pkg/config/config.go` — add `BundlesConfig`, wire into `Config`
- `pkg/http/api/config/` — new package: `init.go` + `config.go`
- `pkg/http/api/api.go` — register `/api/config` subrouter
- `README.md` — add `bundles` config section

---

## sql/bundles_struct.sql

```sql
DROP SCHEMA IF EXISTS bundles CASCADE;

CREATE SCHEMA bundles;
GRANT USAGE ON SCHEMA bundles
    TO edadmin, edservice, edviewer;

CREATE TABLE bundles.bundles (
    id              int          GENERATED ALWAYS AS IDENTITY,
    measurementtype varchar(32)  NOT NULL,
    name            varchar(128) NOT NULL,
    filename        varchar(256) NOT NULL,
    generatedat     timestamp,
    autoregen       bool         NOT NULL DEFAULT false,
    status          varchar(16)  NOT NULL DEFAULT 'pending',
    errormessage    text,
    PRIMARY KEY (id),
    UNIQUE (filename),
    CHECK (status IN (
        'pending', 'queued', 'generating', 'ready', 'error'
    ))
);
GRANT SELECT, INSERT, UPDATE, DELETE
    ON bundles.bundles TO edservice;
GRANT SELECT ON bundles.bundles TO edviewer;

CREATE TABLE bundles.vsds_bundles (
    bundleid    int         NOT NULL,
    subtype     varchar(32) NOT NULL,
    allprojects bool        NOT NULL DEFAULT false,
    PRIMARY KEY (bundleid),
    FOREIGN KEY (bundleid)
        REFERENCES bundles.bundles (id) ON DELETE CASCADE,
    CHECK (subtype IN ('surveypoints', 'surveys'))
);
GRANT SELECT, INSERT, UPDATE, DELETE
    ON bundles.vsds_bundles TO edservice;
GRANT SELECT ON bundles.vsds_bundles TO edviewer;

CREATE TABLE bundles.vsds_bundle_projects (
    bundleid  int NOT NULL,
    projectid int NOT NULL,
    PRIMARY KEY (bundleid, projectid),
    FOREIGN KEY (bundleid)
        REFERENCES bundles.vsds_bundles (bundleid)
        ON DELETE CASCADE,
    FOREIGN KEY (projectid)
        REFERENCES vsds.projects (id) ON DELETE CASCADE
);
GRANT SELECT, INSERT, UPDATE, DELETE
    ON bundles.vsds_bundle_projects TO edservice;
GRANT SELECT ON bundles.vsds_bundle_projects TO edviewer;
```

---

## sql/bundles_views.sql

```sql
CREATE OR REPLACE VIEW bundles.v_bundles AS
SELECT id, measurementtype, name, filename,
       generatedat, autoregen, status, errormessage
FROM bundles.bundles;
GRANT SELECT ON bundles.v_bundles TO edservice;
GRANT SELECT ON bundles.v_bundles TO edviewer;

CREATE OR REPLACE VIEW bundles.v_vsds_bundles AS
SELECT b.id, b.measurementtype, b.name, b.filename,
       b.generatedat, b.autoregen,
       b.status, b.errormessage,
       vb.subtype, vb.allprojects,
       CASE
           WHEN vb.allprojects THEN NULL
           ELSE array_agg(
               vbp.projectid ORDER BY vbp.projectid
           ) FILTER (WHERE vbp.projectid IS NOT NULL)
       END AS projects
FROM bundles.bundles b
JOIN bundles.vsds_bundles vb ON vb.bundleid = b.id
LEFT JOIN bundles.vsds_bundle_projects vbp
    ON vbp.bundleid = b.id
WHERE b.measurementtype = 'vsds'
GROUP BY b.id, b.measurementtype, b.name, b.filename,
         b.generatedat, b.autoregen, b.status,
         b.errormessage, vb.subtype, vb.allprojects;
GRANT SELECT ON bundles.v_vsds_bundles TO edservice;
GRANT SELECT ON bundles.v_vsds_bundles TO edviewer;
```

---

## sql/bundles_funcs.sql

Two functions:

### `bundles.create_vsds_bundle`

Inserts into `bundles.bundles`, `bundles.vsds_bundles`, and optionally
`bundles.vsds_bundle_projects`. Computes filename as
`'vsds-' || id || '.json.gz'` after INSERT. Returns the `v_vsds_bundles`
row for the new bundle.

Parameters: `p_name varchar`, `p_autoregen bool`, `p_subtype varchar`,
`p_allprojects bool`, `p_projects int[]` (ignored when `p_allprojects=true`).

Implementation notes:
- Use `RETURNING id` from the bundles INSERT to compute filename
- `UPDATE bundles.bundles SET filename = 'vsds-' || $id || '.json.gz'`
- Only insert project rows if `NOT p_allprojects`
- Return via `SELECT * FROM bundles.v_vsds_bundles WHERE id = $id`

### `bundles.queue_autoregen_bundles`

Parameter: `p_project_ids int[]`

```sql
UPDATE bundles.bundles b
SET status = 'queued'
FROM bundles.vsds_bundles vb
LEFT JOIN bundles.vsds_bundle_projects vbp
    ON vbp.bundleid = vb.bundleid
WHERE b.id = vb.bundleid
  AND b.measurementtype = 'vsds'
  AND b.autoregen = true
  AND b.status != 'generating'
  AND (
      vb.allprojects = true
      OR vbp.projectid = ANY(p_project_ids)
  );
```

---

## sql/_all.sql

Add after vsds imports (bundles has FK into vsds.projects — must come after):

```
\i bundles_struct.sql
\i bundles_views.sql
\i bundles_funcs.sql
```

---

## pkg/config/config.go

Add `BundlesConfig` struct:

```go
type BundlesConfig struct {
    Path          string        `koanf:"path"`
    Serve         bool          `koanf:"serve"`
    BaseURL       string        `koanf:"baseUrl"`
    CheckInterval time.Duration `koanf:"checkInterval"`
}
```

Add field to `Config`:
```go
Bundles BundlesConfig `koanf:"bundles"`
```

Add default in `ParseConfig`:
```go
Bundles: BundlesConfig{
    CheckInterval: 5 * time.Minute,
},
```

---

## pkg/http/api/config/ (new package)

### config.go

Response type and handler for `GET /api/config` (public endpoint).

```go
type AppConfigResponse struct {
    BundleBaseURL string `json:"bundleBaseUrl"`
}
```

The handler reads `cfg.Bundles.BaseURL`. The config must be made accessible to
this package (package-level var set at init, following existing patterns like
`gcp.ClientEmail()`).

Add an `Init(cfg *config.BundlesConfig)` function in the package that stores
the bundles config for the handler.

### init.go

```go
func InitRoutes(r *mux.Router) error {
    r.Handle("", w.NewAPIHandler().Get(getConfig))
    return nil
}
```

---

## pkg/http/api/api.go

Register the config subrouter:

```go
sr = r.PathPrefix("/config").Subrouter()
if err = apicfg.InitRoutes(sr); err != nil {
    return errors.Join(err, fmt.Errorf("Unable init config endpoints"))
}
```

Also pass `cfg.Bundles` to `apicfg.Init(&cfg.Bundles)` before route registration.
Update `api.Init` signature to accept `*config.Config` instead of only
`*config.OAuth2Config` — or pass `BundlesConfig` as a separate parameter. Follow
the existing pattern and minimise signature churn (add a parameter).

---

## README.md

Add a `bundles` subsection under **Configuration**, after `vsds`:

```yaml
bundles:
  path: /var/lib/edst/bundles  # directory for .json.gz bundle files
  serve: false                 # true = serve /static/ from this backend
  baseUrl: ""                  # URL prefix for download links
                               # empty = frontend defaults to /static/
                               # e.g. https://static.example.com/
  checkInterval: 5m            # optional, default 5m
```

Add a **Production static serving** note (can be a subsection under **Running** or
a new **Deployment** section):

```
For production, bundle files are best served from a dedicated static host or
subdomain rather than the application backend.  Point nginx at the configured
`bundles.path` directory with `gzip_static on;` so pre-compressed `.json.gz`
files are served with the correct `Content-Encoding: gzip` header:

    location / {
        root /var/lib/edst/bundles;
        gzip_static on;
        default_type application/json;
    }

Set `bundles.baseUrl` in the application config to the base URL of that host
(e.g. `https://static.example.com/`) so the frontend constructs correct links.
Leave `bundles.serve: false` when using external static serving.
```

---

## Build verification

Run `gmake build` after all changes. Fix any compile errors before completing
this phase.
