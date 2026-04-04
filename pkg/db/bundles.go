package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// QueryRows executes a named prepared query and scans each result row
// into T via RowToStructByName. It is a package-level generic function
// so that callers can specify their own scan-target type without
// needing to define it in this package.
func QueryRows[T any](
	p *DBPool, query string, args ...any,
) ([]T, error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(p.ctx, query, args...)
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", query).
			Msg("Error while executing query")
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectRows(
		rows, pgx.RowToStructByName[T],
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return nil, err
	}
	return result, nil
}

// Bundle is a row from bundles.v_vsds_bundles. Used as the canonical
// in-memory representation for both the processor and API handlers.
type Bundle struct {
	ID              int        `db:"id" json:"id"`
	MeasurementType string     `db:"measurementtype" json:"measurementtype"`
	Name            string     `db:"name" json:"name"`
	Filename        string     `db:"filename" json:"filename"`
	GeneratedAt     *time.Time `db:"generatedat" json:"generated_at,omitempty"`
	AutoRegen       bool       `db:"autoregen" json:"autoregen"`
	Status          string     `db:"status" json:"status"`
	ErrorMessage    *string    `db:"errormessage" json:"error_message,omitempty"`
	// VSDS-specific (populated from v_vsds_bundles; zero-valued for other types)
	Subtype     *string  `db:"subtype" json:"subtype,omitempty"`
	AllProjects *bool    `db:"allprojects" json:"allprojects,omitempty"`
	Projects    []string `db:"projects" json:"projects,omitempty"`
}

// VSDBundleConfig is the type-specific config loaded by the VSDS runner.
// Returned by GetVSDBundleConfig; passed opaquely through the framework.
type VSDBundleConfig struct {
	BundleID    int
	Subtype     string
	AllProjects bool
	ProjectIDs  []int
}

// ListPendingBundles returns all bundles with status 'pending' or
// 'queued', ordered by ID.
func (p *DBPool) ListPendingBundles() ([]Bundle, error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(p.ctx, "listpendingbundles")
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "listpendingbundles").
			Msg("Error while executing query")
		return nil, err
	}
	defer rows.Close()

	bundles, err := pgx.CollectRows(
		rows, pgx.RowToStructByName[Bundle],
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return nil, err
	}

	return bundles, nil
}

// SetBundleGenerating atomically transitions a bundle to 'generating'.
// Returns ErrAlreadyQueued if the bundle was already claimed by
// another processor instance.
func (p *DBPool) SetBundleGenerating(id int) error {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return err
	}
	defer conn.Release()

	tag, err := conn.Exec(p.ctx, "setbundlegenerating", id)
	if err != nil {
		logger.Error().Err(err).Caller().
			Int("id", id).
			Msg("Error executing setbundlegenerating")
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrAlreadyQueued
	}

	return nil
}

// SetBundleReady marks a bundle as successfully generated.
func (p *DBPool) SetBundleReady(id int) error {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(p.ctx, "setbundleready", id)
	if err != nil {
		logger.Error().Err(err).Caller().
			Int("id", id).
			Msg("Error executing setbundleready")
		return err
	}

	return nil
}

// SetBundleError marks a bundle as failed with an error message.
func (p *DBPool) SetBundleError(id int, msg string) error {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(p.ctx, "setbundleerror", id, msg)
	if err != nil {
		logger.Error().Err(err).Caller().
			Int("id", id).
			Msg("Error executing setbundleerror")
		return err
	}

	return nil
}

// GetVSDBundleConfig loads the type-specific configuration for a
// VSDS bundle by ID.
func (p *DBPool) GetVSDBundleConfig(id int) (*VSDBundleConfig, error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return nil, err
	}
	defer conn.Release()

	type cfgRow struct {
		Subtype     string `db:"subtype"`
		AllProjects bool   `db:"allprojects"`
	}

	rows, err := conn.Query(p.ctx, "getvsdsbundleconfig", id)
	if err != nil {
		logger.Error().Err(err).Caller().
			Int("id", id).
			Str("query", "getvsdsbundleconfig").
			Msg("Error while executing query")
		return nil, err
	}
	defer rows.Close()

	cfgRows, err := pgx.CollectRows(
		rows, pgx.RowToStructByName[cfgRow],
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading config results")
		return nil, err
	}

	if len(cfgRows) == 0 {
		return nil, ErrNotFound
	}

	cfg := &VSDBundleConfig{
		BundleID:    id,
		Subtype:     cfgRows[0].Subtype,
		AllProjects: cfgRows[0].AllProjects,
	}

	projRows, err := conn.Query(
		p.ctx, "getvsdsbundleprojects", id,
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Int("id", id).
			Str("query", "getvsdsbundleprojects").
			Msg("Error while executing query")
		return nil, err
	}
	defer projRows.Close()

	type projRow struct {
		ProjectID int `db:"projectid"`
	}

	prows, err := pgx.CollectRows(
		projRows, pgx.RowToStructByName[projRow],
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading project results")
		return nil, err
	}

	for _, pr := range prows {
		cfg.ProjectIDs = append(cfg.ProjectIDs, pr.ProjectID)
	}

	return cfg, nil
}

// ListBundles returns all bundles ordered by ID.
func (p *DBPool) ListBundles() ([]Bundle, error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(p.ctx, "listbundles")
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "listbundles").
			Msg("Error while executing query")
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectRows(
		rows, pgx.RowToStructByName[Bundle],
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return nil, err
	}
	return result, nil
}

// GetBundle returns a bundle by ID.
// Returns ErrNotFound when no bundle with that ID exists.
func (p *DBPool) GetBundle(id int) (*Bundle, error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(p.ctx, "getbundle", id)
	if err != nil {
		logger.Error().Err(err).Caller().
			Int("id", id).
			Str("query", "getbundle").
			Msg("Error while executing query")
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectRows(
		rows, pgx.RowToStructByName[Bundle],
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return nil, err
	}

	if len(result) == 0 {
		return nil, ErrNotFound
	}
	return &result[0], nil
}

// CreateVSDSBundle creates a new VSDS bundle.
// When allprojects is true, projects is ignored.
// Returns the newly created bundle row.
func (p *DBPool) CreateVSDSBundle(
	name string,
	autoregen bool,
	subtype string,
	allprojects bool,
	projects []int,
) (*Bundle, error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return nil, err
	}
	defer conn.Release()

	proj32 := make([]int32, len(projects))
	for i, id := range projects {
		proj32[i] = int32(id)
	}

	rows, err := conn.Query(
		p.ctx, "createvsdsbundle",
		name, autoregen, subtype, allprojects, proj32,
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error executing createvsdsbundle")
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectRows(
		rows, pgx.RowToStructByName[Bundle],
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error reading created bundle")
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf(
			"create_vsds_bundle returned no row")
	}
	return &result[0], nil
}

// DeleteBundle deletes a bundle by ID.
// Returns ErrNotFound when no bundle with that ID exists.
func (p *DBPool) DeleteBundle(id int) error {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return err
	}
	defer conn.Release()

	tag, err := conn.Exec(p.ctx, "deletebundle", id)
	if err != nil {
		logger.Error().Err(err).Caller().
			Int("id", id).
			Msg("Error executing deletebundle")
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// QueueBundle sets a bundle's status to 'queued'.
// Returns ErrAlreadyQueued when the bundle is currently generating.
// Returns ErrNotFound when no bundle with that ID exists.
func (p *DBPool) QueueBundle(id int) error {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return err
	}
	defer conn.Release()

	tag, err := conn.Exec(p.ctx, "queuebundle", id)
	if err != nil {
		logger.Error().Err(err).Caller().
			Int("id", id).
			Msg("Error executing queuebundle")
		return err
	}

	if tag.RowsAffected() == 0 {
		var dummy int
		err = conn.QueryRow(
			p.ctx, "checkbundleexists", id,
		).Scan(&dummy)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			logger.Error().Err(err).Caller().
				Int("id", id).
				Msg("Error checking bundle existence")
			return err
		}
		return ErrAlreadyQueued
	}
	return nil
}

// UpdateBundleAutoregen updates the autoregen flag for a bundle.
// Returns ErrNotFound when no bundle with that ID exists.
func (p *DBPool) UpdateBundleAutoregen(
	id int, autoregen bool,
) (*Bundle, error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return nil, err
	}
	defer conn.Release()

	tag, err := conn.Exec(
		p.ctx, "updatebundleautoregen", id, autoregen,
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Int("id", id).
			Msg("Error executing updatebundleautoregen")
		return nil, err
	}

	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	return p.GetBundle(id)
}

// UpdateVSDSBundle conditionally updates mutable fields of a VSDS
// bundle. Only non-nil parameters are applied. Changing vsds-specific
// fields (subtype, allprojects, projects) resets status to 'pending'.
// Returns ErrNotFound when no bundle with that ID exists.
func (p *DBPool) UpdateVSDSBundle(
	id int,
	name *string,
	autoregen *bool,
	subtype *string,
	allprojects *bool,
	projects []int,
) (*Bundle, error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return nil, err
	}
	defer conn.Release()

	tx, err := conn.Begin(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error starting transaction")
		return nil, err
	}
	defer tx.Rollback(p.ctx)

	var curAllProjects bool
	err = tx.QueryRow(
		p.ctx,
		`SELECT allprojects FROM bundles.vsds_bundles
         WHERE bundleid = $1`,
		id,
	).Scan(&curAllProjects)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		logger.Error().Err(err).Caller().
			Int("id", id).
			Msg("Error reading vsds_bundles row")
		return nil, err
	}

	if name != nil {
		_, err = tx.Exec(
			p.ctx,
			`UPDATE bundles.bundles
             SET name = $2 WHERE id = $1`,
			id, *name,
		)
		if err != nil {
			logger.Error().Err(err).Caller().
				Int("id", id).
				Msg("Error updating bundle name")
			return nil, err
		}
	}

	if autoregen != nil {
		_, err = tx.Exec(
			p.ctx,
			`UPDATE bundles.bundles
             SET autoregen = $2 WHERE id = $1`,
			id, *autoregen,
		)
		if err != nil {
			logger.Error().Err(err).Caller().
				Int("id", id).
				Msg("Error updating bundle autoregen")
			return nil, err
		}
	}

	if subtype != nil {
		_, err = tx.Exec(
			p.ctx,
			`UPDATE bundles.vsds_bundles
             SET subtype = $2 WHERE bundleid = $1`,
			id, *subtype,
		)
		if err != nil {
			logger.Error().Err(err).Caller().
				Int("id", id).
				Msg("Error updating vsds_bundles subtype")
			return nil, err
		}
	}

	if allprojects != nil {
		_, err = tx.Exec(
			p.ctx,
			`UPDATE bundles.vsds_bundles
             SET allprojects = $2 WHERE bundleid = $1`,
			id, *allprojects,
		)
		if err != nil {
			logger.Error().Err(err).Caller().
				Int("id", id).
				Msg("Error updating vsds_bundles allprojects")
			return nil, err
		}
	}

	vsdsChanged := subtype != nil ||
		allprojects != nil || projects != nil
	if vsdsChanged {
		effAllProjects := curAllProjects
		if allprojects != nil {
			effAllProjects = *allprojects
		}

		_, err = tx.Exec(
			p.ctx,
			`DELETE FROM bundles.vsds_bundle_projects
             WHERE bundleid = $1`,
			id,
		)
		if err != nil {
			logger.Error().Err(err).Caller().
				Int("id", id).
				Msg("Error deleting vsds_bundle_projects")
			return nil, err
		}

		if projects != nil && !effAllProjects {
			for _, pid := range projects {
				_, err = tx.Exec(
					p.ctx,
					`INSERT INTO bundles.vsds_bundle_projects
                     (bundleid, projectid)
                     VALUES ($1, $2)`,
					id, int32(pid),
				)
				if err != nil {
					logger.Error().Err(err).Caller().
						Int("id", id).
						Int("projectid", pid).
						Msg("Error inserting vsds_bundle_projects")
					return nil, err
				}
			}
		}

		_, err = tx.Exec(
			p.ctx,
			`UPDATE bundles.bundles
             SET status = 'pending' WHERE id = $1`,
			id,
		)
		if err != nil {
			logger.Error().Err(err).Caller().
				Int("id", id).
				Msg("Error resetting bundle status to pending")
			return nil, err
		}
	}

	if err = tx.Commit(p.ctx); err != nil {
		logger.Error().Err(err).Caller().
			Int("id", id).
			Msg("Error committing UpdateVSDSBundle transaction")
		return nil, err
	}

	return p.GetBundle(id)
}

// QueueAutoRegenBundles sets status='queued' for all autoregen bundles
// whose project scope intersects projectIDs. No-op when projectIDs
// is empty.
func (p *DBPool) QueueAutoRegenBundles(projectIDs []int) error {
	if len(projectIDs) == 0 {
		return nil
	}

	ids32 := make([]int32, len(projectIDs))
	for i, id := range projectIDs {
		ids32[i] = int32(id)
	}

	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(p.ctx, "queueautoregen", ids32)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error executing queueautoregen")
		return err
	}

	return nil
}
