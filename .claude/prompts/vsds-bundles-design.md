# VSDS Bundles — Design Document

**State**: ACCEPTED. Use this as the reference for all phase implementation prompts.

---

## DB Schema — IS-A pattern

Separate `bundles` schema. Generic fields in base table; type-specific in subtype tables.
Cross-schema FKs are allowed and used.

**`bundles.bundles`** (supertype):
```
id              int IDENTITY PK
measurementtype varchar(32) NOT NULL
name            varchar(128) NOT NULL
filename        varchar(256) NOT NULL UNIQUE
generatedat     timestamp
autoregen       bool NOT NULL DEFAULT false
status          varchar(16) NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending','queued','generating','ready','error'))
errormessage    text
```

**`bundles.vsds_bundles`** (VSDS subtype, 1:1):
```
bundleid    int PK, FK → bundles.bundles(id) ON DELETE CASCADE
subtype     varchar(32) NOT NULL
            CHECK (subtype IN ('surveypoints', 'surveys'))
allprojects bool NOT NULL DEFAULT false
```

**`bundles.vsds_bundle_projects`** (VSDS project scope):
```
bundleid  int FK → bundles.vsds_bundles(bundleid) ON DELETE CASCADE
projectid int NOT NULL, FK → vsds.projects(id) ON DELETE CASCADE
PK (bundleid, projectid)
```

`vsds_bundle_projects` is explicitly VSDS-specific. It is only JOINed when
`measurementtype='vsds'`. The FK to `vsds.projects` is correct — the IS-A structure
guarantees VSDS context by the time this table is reached.

Views:
- `bundles.v_bundles`: base view of `bundles.bundles`
- `bundles.v_vsds_bundles`: full join of all three VSDS tables; aggregates `projectid`s
  into array (NULL when `allprojects=true`)

DB functions:
- `bundles.create_vsds_bundle(...)`: INSERT into bundles + vsds_bundles + projects,
  computes filename from generated ID, returns v_vsds_bundles row
- `bundles.queue_autoregen_bundles(p_project_ids int[])`: UPDATE status='queued' for
  autoregen VSDS bundles whose project scope intersects the given IDs

A future `stellar_bundles` adds its own subtype table under `bundles.` without touching
the base table.

Roles: `edservice` SELECT/INSERT/UPDATE/DELETE on all tables; `edviewer` SELECT only.

---

## Filename — ID-based, stable

`{measurementtype}-{id}.json.gz` — e.g. `vsds-42.json.gz`

Set by the DB creation function after ID is generated. Never changes. Regeneration
overwrites in-place. No hash, no loitering files, no broken links on config edits.

---

## Status machine

```
pending → queued → generating → ready
                              ↘ error
ready / error → queued (via regen trigger)
```

- `pending`: created, never generated (also reset when type-specific config is PATCHed)
- `queued`: generation requested, not yet picked up
- `generating`: processor is currently running this bundle
- `ready`: file is valid and up-to-date
- `error`: last attempt failed; old file (if any) remains untouched

Regen trigger (auto or manual): sets `status='queued'` when not `'generating'`; 409 otherwise.
PATCH on type-specific fields resets `status='pending'` (existing file marked stale).

---

## Config additions

New `BundlesConfig` in `pkg/config/config.go`:

```go
type BundlesConfig struct {
    Path          string        `koanf:"path"`
    Serve         bool          `koanf:"serve"`
    BaseURL       string        `koanf:"baseUrl"`
    CheckInterval time.Duration `koanf:"checkInterval"`
}
```

`BaseURL` accepts relative (`/static/`) or absolute (`https://static.example.com/`).
Empty → frontend defaults to `/static/`.

Exposed via a new generic `/api/config` endpoint (extensible; auth endpoints stay
separate).

README.md must be updated with the `bundles` config section and the nginx `gzip_static`
deployment note.

---

## Bundle Processor — `pkg/bundles/processor.go`

Patterned after `pkg/vsds/processor.go`, but persistent — loops until stopped:

```go
type Processor struct {
    cfg    *config.BundlesConfig
    logger log.Logger
    stopCh chan struct{}
    doneCh chan struct{}
    sigCh  chan struct{}
}

var P *Processor  // package-level instance, set by service.go

func Signal() {
    if P == nil { return }
    select { case P.sigCh <- struct{}{}: default: }
}
```

`run()`: calls `process()` immediately on start, then waits on stopCh / sigCh / ticker.

`process()`:
1. Query `bundles.bundles WHERE status IN ('pending','queued')`
2. For each bundle, sequentially:
   a. Atomically set `status='generating'`
   b. Look up runner in registry by `measurementtype`
   c. `cfg := runner.LoadConfig(bundleID)`
   d. `data := runner.Generate(cfg)` → returns structured `any`
   e. Framework: JSON marshal → gzip → atomic write (temp file + rename)
   f. Set `status='ready'` + `generatedat=now()` or `status='error'` + `errormessage`

---

## Generator Registry — `pkg/bundles/generator.go`

```go
type BundleRunner interface {
    LoadConfig(bundleID int) (any, error)
    Generate(config any) (any, error)
}

var registry = map[string]BundleRunner{}

func Register(measurementType string, runner BundleRunner) {
    registry[measurementType] = runner
}
```

Framework owns all I/O (JSON serialize, gzip compress, atomic file write).
Runners own data fetching and structuring only.

Registration in `cmd/edst/service.go`:
```go
bundles.Register("vsds", vsds.NewBundleRunner())
```

**VSDS runner** (`pkg/vsds/bundlerunner.go`):
- `LoadConfig(id)` → queries `bundles.vsds_bundles` + `bundles.vsds_bundle_projects`
- `Generate(cfg)` → queries `vsds.v_surveypoints` or `vsds.v_surveys` based on subtype,
  filtered by project IDs (or unfiltered for `allprojects=true`)
- Output structs strip internal IDs (`id`, `surveyid` not serialised)
- `v_surveys` view shape used as-is for surveys bundle (known bugs fixed separately)

---

## Auto-regen trigger from VSDS processor

In `pkg/vsds/processor.process()`, collect affected project IDs throughout the run.
The defer block (after `FinishFolderProcessing` and `RefreshSurveyMaterializedViews`) adds:

```go
if err := db.Pool.QueueAutoRegenBundles(affectedProjectIDs); err != nil {
    p.logger.Error().Err(err).Msg("Error queuing auto-regen bundles")
}
bundles.Signal()
```

`QueueAutoRegenBundles` only queues bundles whose project scope intersects
`affectedProjectIDs` (or `allprojects=true`).

---

## API — `pkg/http/api/bundles/`

| Method | Path | Auth | Notes |
|---|---|---|---|
| `GET` | `/api/bundles` | Public | List all |
| `GET` | `/api/bundles/{id}` | Public | Full detail incl. type-specific fields |
| `PUT` | `/api/bundles` | Admin | Create |
| `DELETE` | `/api/bundles/{id}` | Admin | DB row only; no file deletion |
| `POST` | `/api/bundles/{id}/generate` | Admin | Set queued + Signal(); 409 if generating |
| `PATCH` | `/api/bundles/{id}` | Admin | Update any mutable fields |

Shared PUT/PATCH body structure:
```json
{
    "measurementtype": "vsds",
    "name": "DW3 Full Survey Points",
    "autoregen": true,
    "vsds": {
        "subtype": "surveypoints",
        "allprojects": false,
        "projects": [1, 2]
    }
}
```

`measurementtype` immutable after creation. PATCH on vsds-specific fields resets
`status='pending'`.

---

## Serving `/static/`

Custom handler in `pkg/http/http.go`, registered before the SPA catch-all when
`cfg.Bundles.Serve == true`. Sets `Content-Type: application/json` and
`Content-Encoding: gzip` for `.json.gz` files. Browser and Angular `HttpClient`
decompress transparently.

For prod nginx deployment: add `gzip_static on;` in the location block serving the
bundles path on the `static.` subdomain. Document this in README.md.

---

## Housekeeping (orphaned files)

Deferred. Future `isOwner`-gated sweep: scan `bundles.path`, cross-reference
`bundles.bundles.filename`, remove orphans.

---

## Frontend — top-level `Bundles` section

- **Public** top-level navbar section (visible to all including anonymous)
- Landing page: subsections carousel
- **VSDS Bundles** subsection (`/bundles/vsds`):
  - Public view: status, last generated, Download link (`{bundleBaseUrl}/{filename}`)
  - When `isAdmin`: same page extended with name, subtype, projects, autoregen toggle,
    Generate / Delete / Add buttons
  - Add-bundle form: name, subtype dropdown, project multi-select / All Projects
    checkbox, autoregen checkbox
- New `BundlesService` (`src/app/services/bundles.service.ts`)
- `bundleBaseUrl` from `/api/config` on startup

---

## Implementation phases

Phase files at `.claude/prompts/vsds-bundles-phaseN.md`.

| Phase | Scope |
|---|---|
| 1 | DB schema + `BundlesConfig` + `/api/config` endpoint + README config section |
| 2 | `pkg/bundles/` framework: processor, runner registry, I/O writer, DB methods |
| 3 | VSDS runner + auto-regen trigger in VSDS processor |
| 4 | REST API (`pkg/http/api/bundles/`), `/static/` handler, README nginx note |
| 5 | Frontend: Bundles section, VSDS subsection, BundlesService |

---

## Resolved design decisions

| # | Question | Decision |
|---|---|---|
| 1 | Sequential vs parallel in `process()` | Sequential |
| 2 | Auto-regen scope | Fine-grained — intersect with affected project IDs |
| 3 | DB schema placement | `bundles.` schema; cross-schema FKs to `vsds.projects` allowed |
| 4 | Surveys bundle shape | Structured per `v_surveys` view; view bugs fixed separately |
