# Phase 2 — Bundle Framework: Processor, Registry, I/O

This implementation has been previously designed and approved.
Proceed with implementation directly.

Reference: `.claude/prompts/vsds-bundles-design.md`
Prerequisite: Phase 1 complete (DB schema, BundlesConfig in config).

---

## Scope

- `pkg/db/bundles.go` — DB types + methods used by the processor
- `pkg/bundles/generator.go` — `BundleRunner` interface + registry
- `pkg/bundles/writer.go` — JSON marshal, gzip compress, atomic file write
- `pkg/bundles/processor.go` — `Processor`, package-level `P`, `Signal()`, run loop
- `cmd/edst/service.go` — instantiate + start + stop the bundle processor

No VSDS runner registration yet — that is Phase 3.

---

## pkg/db/bundles.go

### Types

```go
// Bundle is a row from bundles.v_vsds_bundles. Used as the canonical
// in-memory representation for both the processor and API handlers.
type Bundle struct {
    ID              int        `db:"id"              json:"id"`
    MeasurementType string     `db:"measurementtype" json:"measurementtype"`
    Name            string     `db:"name"            json:"name"`
    Filename        string     `db:"filename"        json:"filename"`
    GeneratedAt     *time.Time `db:"generatedat"     json:"generated_at,omitempty"`
    AutoRegen       bool       `db:"autoregen"       json:"autoregen"`
    Status          string     `db:"status"          json:"status"`
    ErrorMessage    *string    `db:"errormessage"    json:"error_message,omitempty"`
    // VSDS-specific (populated from v_vsds_bundles; zero-valued for other types)
    Subtype         *string    `db:"subtype"         json:"subtype,omitempty"`
    AllProjects     *bool      `db:"allprojects"     json:"allprojects,omitempty"`
    Projects        []int32    `db:"projects"        json:"projects,omitempty"`
}

// VSDBundleConfig is the type-specific config loaded by the VSDS runner.
// Returned by GetVSDBundleConfig; passed opaquely through the framework.
type VSDBundleConfig struct {
    BundleID    int
    Subtype     string
    AllProjects bool
    ProjectIDs  []int
}
```

### Methods

**`(p *DBPool) ListPendingBundles() ([]Bundle, error)`**
```sql
SELECT * FROM bundles.v_vsds_bundles
WHERE status IN ('pending', 'queued')
ORDER BY id
```
(Extend to other types when they are added — for now only vsds view exists.)

**`(p *DBPool) SetBundleGenerating(id int) error`**

Atomic: only transitions from `pending`/`queued`. If already `generating`, no-op
(another processor instance may have picked it up — treat as ErrAlreadyQueued or
just ignore the 0-rows-updated case).
```sql
UPDATE bundles.bundles
SET status = 'generating'
WHERE id = $1 AND status IN ('pending', 'queued')
```
Check `RowsAffected()`. If 0, return a sentinel error (e.g. `ErrAlreadyQueued`).

**`(p *DBPool) SetBundleReady(id int) error`**
```sql
UPDATE bundles.bundles
SET status = 'ready', generatedat = now(), errormessage = NULL
WHERE id = $1
```

**`(p *DBPool) SetBundleError(id int, msg string) error`**
```sql
UPDATE bundles.bundles
SET status = 'error', errormessage = $2
WHERE id = $1
```

**`(p *DBPool) GetVSDBundleConfig(id int) (*VSDBundleConfig, error)`**

Two queries (or one join):
1. `SELECT subtype, allprojects FROM bundles.vsds_bundles WHERE bundleid = $1`
2. `SELECT projectid FROM bundles.vsds_bundle_projects WHERE bundleid = $1 ORDER BY projectid`

Return `*VSDBundleConfig`.

**`(p *DBPool) QueueAutoRegenBundles(projectIDs []int) error`**

Calls `SELECT bundles.queue_autoregen_bundles($1::int[])`.
Pass projectIDs as a pgx array parameter.
If `len(projectIDs) == 0`, skip (no projects affected, nothing to queue).

---

## pkg/bundles/generator.go

```go
package bundles

// BundleRunner is implemented by each measurement-type subsystem.
// The framework calls LoadConfig then Generate; all I/O is handled
// by the framework.
type BundleRunner interface {
    // LoadConfig fetches type-specific configuration for the given
    // bundle ID from the database.
    LoadConfig(bundleID int) (any, error)

    // Generate builds the data structure to be serialised. The config
    // argument is the value returned by LoadConfig.
    Generate(config any) (any, error)
}

var registry = map[string]BundleRunner{}

// Register associates a runner with a measurement type. Call from
// service.go before starting the processor.
func Register(measurementType string, runner BundleRunner) {
    registry[measurementType] = runner
}

func get(measurementType string) (BundleRunner, bool) {
    r, ok := registry[measurementType]
    return r, ok
}
```

---

## pkg/bundles/writer.go

Handles JSON → gzip → atomic file write. Used exclusively by `process()`.

```go
package bundles

// write serialises data to JSON, gzip-compresses it, and writes it
// atomically to destPath (temp file in the same directory + rename).
func write(data any, destPath string) error
```

Implementation:
1. Create a temp file in `filepath.Dir(destPath)` with `os.CreateTemp`
2. Wrap with `gzip.NewWriter`
3. `json.NewEncoder(gz).Encode(data)`
4. Close gzip writer, then close the temp file
5. `os.Rename(tmpFile.Name(), destPath)`
6. On any error, remove the temp file before returning

---

## pkg/bundles/processor.go

```go
package bundles

import (
    "path/filepath"
    "time"

    "github.com/gczuczy/ed-survey-tools/pkg/config"
    "github.com/gczuczy/ed-survey-tools/pkg/db"
    "github.com/gczuczy/ed-survey-tools/pkg/log"
)

// P is the package-level processor instance, set by service.go.
var P *Processor

// Signal wakes the processor immediately to check for queued bundles.
// Safe to call before Start(); the signal is discarded if P is nil.
func Signal() {
    if P == nil {
        return
    }
    select {
    case P.sigCh <- struct{}{}:
    default:
    }
}

type Processor struct {
    cfg    *config.BundlesConfig
    logger log.Logger
    stopCh chan struct{}
    doneCh chan struct{}
    sigCh  chan struct{}
}

func NewProcessor(cfg *config.BundlesConfig) *Processor {
    return &Processor{
        cfg:    cfg,
        logger: log.GetLogger("BundleProcessor"),
        stopCh: make(chan struct{}),
        doneCh: make(chan struct{}),
        sigCh:  make(chan struct{}, 1),
    }
}

func (p *Processor) Start() {
    go p.run()
}

func (p *Processor) Stop() {
    close(p.stopCh)
    <-p.doneCh
}

func (p *Processor) run() {
    defer func() {
        p.logger.Info().Msg("Processor stopped")
        close(p.doneCh)
    }()
    p.logger.Info().Msg("Processor started")

    p.process()

    ticker := time.NewTicker(p.cfg.CheckInterval)
    defer ticker.Stop()
    for {
        select {
        case <-p.stopCh:
            return
        case <-p.sigCh:
            p.process()
        case <-ticker.C:
            p.process()
        }
    }
}

func (p *Processor) process() {
    bundles, err := db.Pool.ListPendingBundles()
    if err != nil {
        p.logger.Error().Err(err).
            Msg("Error listing pending bundles")
        return
    }

    for _, b := range bundles {
        p.generate(b)
    }
}

func (p *Processor) generate(b db.Bundle) {
    if err := db.Pool.SetBundleGenerating(b.ID); err != nil {
        // Another goroutine claimed it; skip silently.
        return
    }

    runner, ok := get(b.MeasurementType)
    if !ok {
        p.logger.Error().
            Int("id", b.ID).
            Str("type", b.MeasurementType).
            Msg("No runner registered for measurement type")
        _ = db.Pool.SetBundleError(b.ID,
            "no runner for type: "+b.MeasurementType)
        return
    }

    cfg, err := runner.LoadConfig(b.ID)
    if err != nil {
        p.logger.Error().Err(err).
            Int("id", b.ID).Msg("Error loading bundle config")
        _ = db.Pool.SetBundleError(b.ID, err.Error())
        return
    }

    data, err := runner.Generate(cfg)
    if err != nil {
        p.logger.Error().Err(err).
            Int("id", b.ID).Msg("Error generating bundle")
        _ = db.Pool.SetBundleError(b.ID, err.Error())
        return
    }

    destPath := filepath.Join(p.cfg.Path, b.Filename)
    if err = write(data, destPath); err != nil {
        p.logger.Error().Err(err).
            Int("id", b.ID).
            Str("path", destPath).
            Msg("Error writing bundle file")
        _ = db.Pool.SetBundleError(b.ID, err.Error())
        return
    }

    if err = db.Pool.SetBundleReady(b.ID); err != nil {
        p.logger.Error().Err(err).
            Int("id", b.ID).Msg("Error marking bundle ready")
    }

    p.logger.Info().
        Int("id", b.ID).
        Str("file", b.Filename).
        Msg("Bundle generated")
}
```

Line-length limit: 82 chars. Adjust wrapping as needed.

---

## cmd/edst/service.go

Add import for `pkg/bundles`. After existing processor setup:

```go
bundles.P = bundles.NewProcessor(&cfg.Bundles)
bundles.P.Start()
```

In the shutdown block, after `proc.Stop()`:

```go
bundles.P.Stop()
```

No runner registrations yet (Phase 3). The processor will start but find no
runners; any pending bundles at startup will log "no runner" errors until
Phase 3 is complete. That is acceptable.

---

## Build verification

Run `gmake build`. Fix all compile errors before completing this phase.
