# Phase 3 — VSDS Runner + Auto-Regen Trigger

This implementation has been previously designed and approved.
Proceed with implementation directly.

Reference: `.claude/prompts/vsds-bundles-design.md`
Prerequisites: Phase 1 and Phase 2 complete.

---

## Scope

- `pkg/vsds/bundlerunner.go` — VSDS `BundleRunner` implementation
- `pkg/vsds/processor.go` — collect affected project IDs; add auto-regen trigger
- `cmd/edst/service.go` — register VSDS runner

No new DB files. `db.GetVSDBundleConfig` and `db.QueueAutoRegenBundles` were added in
Phase 2.

---

## pkg/vsds/bundlerunner.go

New file in package `vsds`.

### Output types

Define two output structs — fields match `vsds.v_surveypoints` and `vsds.v_surveys`
respectively, with internal IDs excluded from JSON.

**Survey point output** (from `v_surveypoints`):

The view columns are: `id`, `surveyid`, `sysname`, `zsample`, `x`, `y`, `z`,
`corrected_n`, `maxdistance`, `rho`, `gc_x`, `gc_y`, `gc_z`.

Strip `id` and `surveyid`; expose `corrected_n` as `"syscount"`:

```go
type bundleSurveyPoint struct {
    SysName     string  `db:"sysname"      json:"sysname"`
    ZSample     int     `db:"zsample"      json:"zsample"`
    X           float32 `db:"x"            json:"x"`
    Y           float32 `db:"y"            json:"y"`
    Z           float32 `db:"z"            json:"z"`
    GCX         float32 `db:"gc_x"         json:"gc_x"`
    GCY         float32 `db:"gc_y"         json:"gc_y"`
    GCZ         float32 `db:"gc_z"         json:"gc_z"`
    SysCount    int     `db:"corrected_n"  json:"syscount"`
    MaxDistance float32 `db:"maxdistance"  json:"maxdistance"`
    Rho         float64 `db:"rho"          json:"rho"`
}
```

**Survey output** (from `v_surveys`):

The view columns are: `projectname`, `id`, `surveyid`, `rho_max`, `x`, `z`,
`rho_stddev`, `gc_x`, `gc_z`, `points` (jsonb).

Strip `id` and `surveyid`. `points` is a `jsonb` column containing
`[{"zsample": N, "rho": F}, ...]` — scan into `json.RawMessage`:

```go
type bundleSurvey struct {
    ProjectName string          `db:"projectname" json:"projectname"`
    RhoMax      float64         `db:"rho_max"     json:"rho_max"`
    X           float64         `db:"x"           json:"x"`
    Z           float64         `db:"z"           json:"z"`
    RhoStddev   *float64        `db:"rho_stddev"  json:"rho_stddev"`
    GCX         float64         `db:"gc_x"        json:"gc_x"`
    GCZ         float64         `db:"gc_z"        json:"gc_z"`
    Points      json.RawMessage `db:"points"      json:"points"`
}
```

### VSDSBundleRunner

```go
type VSDSBundleRunner struct{}

func NewBundleRunner() *VSDSBundleRunner {
    return &VSDSBundleRunner{}
}

func (r *VSDSBundleRunner) LoadConfig(
    bundleID int,
) (any, error) {
    return db.Pool.GetVSDBundleConfig(bundleID)
}

func (r *VSDSBundleRunner) Generate(config any) (any, error) {
    cfg, ok := config.(*db.VSDBundleConfig)
    if !ok {
        return nil, fmt.Errorf(
            "unexpected config type %T", config)
    }
    switch cfg.Subtype {
    case "surveypoints":
        return r.generateSurveyPoints(cfg)
    case "surveys":
        return r.generateSurveys(cfg)
    default:
        return nil, fmt.Errorf(
            "unknown vsds subtype: %s", cfg.Subtype)
    }
}
```

### generateSurveyPoints

Query `vsds.v_surveypoints`. When `cfg.AllProjects` is false, filter by
`surveyid IN (SELECT id FROM vsds.surveys WHERE projectid = ANY($1::int[]))`.
When true, no filter.

Use `pgx.CollectRows` / `pgx.RowToStructByName` (matching existing patterns in
`pkg/db/`). Return `[]bundleSurveyPoint`.

### generateSurveys

Same pattern for `vsds.v_surveys`. Return `[]bundleSurvey`.

`points` is a jsonb column. Use `pgtype.Text` or scan as `[]byte` then wrap in
`json.RawMessage`. Verify the scan approach matches what pgx provides for jsonb.
Use `go doc github.com/jackc/pgx/v5/pgtype` if needed.

---

## pkg/vsds/processor.go — modifications

### 1. Track affected project IDs

Add a local `affectedProjects` set at the top of `process()`:

```go
affectedProjects := make(map[int]struct{})
```

After the successful `txn.LookupProject(survey.Project)` call (where `projectID`
is returned), add:

```go
affectedProjects[projectID] = struct{}{}
```

### 2. Auto-regen trigger in defer

The existing `defer` block currently calls `FinishFolderProcessing` and
`RefreshSurveyMaterializedViews`. Add the trigger after both, still inside the
same `defer func()`:

```go
// Collect affected project IDs for auto-regen.
if len(affectedProjects) > 0 {
    pids := make([]int, 0, len(affectedProjects))
    for pid := range affectedProjects {
        pids = append(pids, pid)
    }
    if err := db.Pool.QueueAutoRegenBundles(pids); err != nil {
        p.logger.Error().Err(err).
            Msg("Error queuing auto-regen bundles")
    }
    bundles.Signal()
}
```

Add import for `pkg/bundles`.

The `defer` captures `affectedProjects` by reference (it's a map declared in the
enclosing `process()` scope), so it sees the fully-populated set when it runs.

---

## cmd/edst/service.go

Register the VSDS runner before starting the processor:

```go
bundles.Register("vsds", vsds.NewBundleRunner())
```

Place this line after `bundles.P = bundles.NewProcessor(...)` and before
`bundles.P.Start()`.

---

## Verification notes

- `go doc` the pgx `json.RawMessage` / jsonb scanning behaviour before
  implementing the `points` field scan in `generateSurveys`.
- After build succeeds (`gmake build`), the processor should pick up any bundles
  with `status IN ('pending','queued')` at startup and generate them. Test with a
  manually inserted bundle row if needed.
- Confirm `affectedProjects` is populated correctly: it should only contain IDs
  from surveys that were successfully committed (i.e. after `txn.AddSurvey`
  succeeds and before `recordResult` with `success=false`). Review the existing
  flow carefully — only add to the set when no error path is taken for that sheet.

---

## Build verification

Run `gmake build`. Fix all compile errors before completing this phase.
