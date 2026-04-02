package db

import (
	"fmt"
	"time"
	"errors"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	//"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"

	"github.com/gczuczy/ed-survey-tools/pkg/config"
	"github.com/gczuczy/ed-survey-tools/pkg/log"
)

var (
	Pool *DBPool=nil

	ErrNotFound             = fmt.Errorf("not found")
	ErrDuplicate            = fmt.Errorf("already exists")
	ErrAlreadyQueued        = fmt.Errorf("processing already queued or in progress")
	ErrCustomerIDRequired   = fmt.Errorf(
		"admin requires a linked customer account")

	prepared = map[string]string{
		// logincmdr
		"logincmdr": `
SELECT * FROM common.logincmdr($1::text, $2::bigint)
`,

		"setsystem": `
INSERT INTO common.systems (edsmid, name, x, y, z)
VALUES ($1::bigint, $2::text, $3::float, $4::float, $5::float)
ON CONFLICT (edsmid) DO UPDATE SET edsmid = EXCLUDED.edsmid
RETURNING *
`,
		// add a spreadsheet file record, returns new id
		"addspreadsheet": `
INSERT INTO vsds.spreadsheets (folderid, gcpid, name, contenttype)
VALUES ($1::int, $2::text, $3::text, $4::text)
RETURNING id
`,

		// add a sheet (tab or implicit CSV sheet), returns new id
		"addsheet": `
INSERT INTO vsds.sheets (spreadsheetid, name)
VALUES ($1::int, $2::text)
RETURNING id
`,

		// upsert a CMDR by name, returns id
		"upsertcmdr": `
INSERT INTO common.cmdrs (name) VALUES ($1::text)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id
`,

		// look up a project by name, returns id
		"lookupproject": `
SELECT id FROM vsds.projects WHERE name = $1::text
`,

		// add a survey, returns new id
		"addsurvey": `
INSERT INTO vsds.surveys (projectid, cmdrid, sheetid)
VALUES ($1::int, $2::int, $3::int)
RETURNING id
`,

		// add a survey point
		"addsurveypoint": `
INSERT INTO vsds.surveypoints
    (surveyid, sysid, zsample, syscount, maxdistance)
VALUES ($1::int, $2::bigint, $3::int, $4::int, $5::real)
`,

		// record a sheet processing outcome; $5 = cmdrid (nullable)
		"recordsheetresult": `
INSERT INTO vsds.sheet_processing
    (procid, sheetid, success, message, cmdrid)
VALUES ($1::int, $2::int, $3::boolean,
        NULLIF($4::text, ''), $5::int)
`,

		// VSDS: look up a CMDR id by name (no upsert)
		"lookcmdrbyname": `
SELECT id FROM common.cmdrs WHERE name = $1::text
`,

		// VSDS: CMDR contribution stats from v_cmdr_contribution
		"cmdrsurveystats": `
SELECT surveys, points,
       coldev_min, coldev_avg, coldev_max
FROM vsds.v_cmdr_contribution
WHERE cmdrid = $1::int
`,

		// VSDS: user's attributed failed-sheet rows
		"getusersheeteerrors": `
SELECT doc_id, doc_name, sheet_name, receivedat, message
FROM vsds.v_user_sheet_errors
WHERE cmdrid = $1::int
ORDER BY receivedat DESC, doc_id, sheet_name
`,

		// VSDS: list projects
		"listprojects": `
SELECT id,name,zsamples FROM vsds.v_projects ORDER BY name
`,

		// VSDS: add project
		"addproject": `
INSERT INTO vsds.projects (name)
VALUES ($1::text)
RETURNING id, name, NULL::int[] AS zsamples
`,

		// VSDS: get project by ID
		"getproject": `
SELECT id, name, zsamples FROM vsds.v_projects WHERE id = $1::int
`,

		// VSDS: delete all zsamples for a project
		"deleteprojectzsamples": `
DELETE FROM vsds.project_zsamples WHERE projectid = $1::int
`,

		// VSDS: insert a single zsample for a project
		"insertprojectzsample": `
INSERT INTO vsds.project_zsamples (projectid, zsample) VALUES ($1::int, $2::int)
`,

		// VSDS: delete a single zsample from a project
		"deleteprojectzsample": `
DELETE FROM vsds.project_zsamples
WHERE projectid = $1::int AND zsample = $2::int
`,

		// VSDS: list folders from view
		"listfolders": `
SELECT folderid, name, gcpid, receivedat, startedat, finishedat, inprogress, finished, failed
FROM vsds.v_folders ORDER BY name
`,

		// VSDS: get folder by ID from view
		"getfolder": `
SELECT folderid, name, gcpid, receivedat, startedat, finishedat, inprogress, finished, failed
FROM vsds.v_folders WHERE folderid = $1::int
`,

		// VSDS: insert a folder, returns the new ID
		"addfolder": `
INSERT INTO vsds.folders (gcpid, name) VALUES ($1::text, $2::text) RETURNING id
`,

		// VSDS: delete a folder by ID
		"deletefolder": `
DELETE FROM vsds.folders WHERE id = $1::int
`,

		// VSDS: check if a folder has active or pending processing
		"checkfolderprocessing": `
SELECT EXISTS(SELECT 1 FROM vsds.folder_processing
WHERE folderid = $1::int AND finishedat IS NULL) AS has_active
`,

		// VSDS: insert a folder processing request, returns new ID
		"insertfolderprocessing": `
INSERT INTO vsds.folder_processing (folderid) VALUES ($1::int) RETURNING id
`,

		// VSDS: fetch the oldest unfinished folder processing job, locking it
		"fetchpendingfolderprocessing": `
SELECT fp.id AS procid, f.id AS folderid, f.gcpid
FROM vsds.folder_processing fp
JOIN vsds.folders f ON f.id = fp.folderid
WHERE fp.finishedat IS NULL
ORDER BY fp.receivedat ASC
LIMIT 1
FOR UPDATE OF fp SKIP LOCKED
`,

		// VSDS: mark a folder processing job as started
		"startfolderprocessing": `
UPDATE vsds.folder_processing SET startedat = NOW() WHERE id = $1::int
`,

		// VSDS: mark a folder processing job as finished
		"finishfolderprocessing": `
UPDATE vsds.folder_processing SET finishedat = NOW() WHERE id = $1::int
`,

		// VSDS: delete all spreadsheets of a folder (cascades to surveys/surveypoints)
		"deletefolderspreadsheets": `
DELETE FROM vsds.spreadsheets WHERE folderid = $1::int
`,

		// VSDS: get the last processing run id and folder name
		"getlastfolderprocessing": `
SELECT fp.id, f.name AS folder_name
FROM vsds.folder_processing fp
JOIN vsds.folders f ON f.id = fp.folderid
WHERE fp.folderid = $1::int
ORDER BY fp.receivedat DESC
LIMIT 1
`,

		// VSDS: summary statistics for a processing run
		"getfolderprocessingsummary": `
SELECT fp.receivedat,
       fp.startedat,
       fp.finishedat,
       COUNT(DISTINCT s.id)    AS documents_total,
       COUNT(DISTINCT sh.id)   AS sheets_total,
       COUNT(DISTINCT sp.sheetid)
           FILTER (WHERE sp.success = true)  AS sheets_success,
       COUNT(DISTINCT sp.sheetid)
           FILTER (WHERE sp.success = false) AS sheets_failed,
       COUNT(DISTINCT sv.id)   AS surveys_ingested,
       COUNT(svp.id)           AS points_ingested,
       COUNT(DISTINCT sv.cmdrid) AS cmdrs_count
FROM vsds.folder_processing fp
LEFT JOIN vsds.spreadsheets s
       ON s.folderid = fp.folderid
LEFT JOIN vsds.sheets sh
       ON sh.spreadsheetid = s.id
LEFT JOIN vsds.sheet_processing sp
       ON sp.sheetid = sh.id AND sp.procid = fp.id
LEFT JOIN vsds.surveys sv ON sv.sheetid = sh.id
LEFT JOIN vsds.surveypoints svp ON svp.surveyid = sv.id
WHERE fp.id = $1::int
GROUP BY fp.id, fp.receivedat, fp.startedat, fp.finishedat
`,

		// VSDS: failing sheets with their document for a processing run
		"getfolderprocessingsheets": `
SELECT s.id AS doc_id,
       s.gcpid,
       s.name AS doc_name,
       s.contenttype,
       sh.id AS sheet_id,
       COALESCE(sh.name, '(CSV sheet)') AS sheet_name,
       COALESCE(sp.message, '') AS message
FROM vsds.spreadsheets s
JOIN vsds.sheets sh ON sh.spreadsheetid = s.id
JOIN vsds.sheet_processing sp ON sp.sheetid = sh.id
WHERE s.folderid = $1::int
  AND sp.procid = $2::int
  AND sp.success = false
ORDER BY s.name, sh.name
`,

		// common: look up systems by name
		"lookupsystemsbyname": `
SELECT id, edsmid, name, x, y, z
FROM common.systems
WHERE name = ANY($1::text[])
`,

		// VSDS: fetch all sheet variants with their project names
		"fetchsheetvariants": `
SELECT sv.id, sv.name, p.name AS projectname,
       sv.headerrow, sv.sysnamecolumn, sv.zsamplecolumn,
       sv.systemcountcolumn, sv.maxdistancecolumn
FROM vsds.spreadsheetvariants sv
JOIN vsds.projects p ON p.id = sv.projectid
ORDER BY sv.id
`,

		// VSDS: fetch all sheet variant header checks
		"fetchsheetvariantchecks": `
SELECT id, variantid, col, row, value
FROM vsds.spreadsheetvariant_checks
ORDER BY variantid, id
`,

		// VSDS: list all variants for a project (with checks)
		"listprojectvariants": `
SELECT id, projectid, projectname, name, headerrow,
       sysnamecolumn, zsamplecolumn,
       systemcountcolumn, maxdistancecolumn,
       checks::text AS checks
FROM vsds.v_spreadsheetvariants
WHERE projectid = $1::int
ORDER BY id
`,

		// VSDS: get one variant by id within a project
		"getvariant": `
SELECT id, projectid, projectname, name, headerrow,
       sysnamecolumn, zsamplecolumn,
       systemcountcolumn, maxdistancecolumn,
       checks::text AS checks
FROM vsds.v_spreadsheetvariants
WHERE id = $1::int AND projectid = $2::int
`,

		// VSDS: insert a new variant, returns new id
		"addvariant": `
INSERT INTO vsds.spreadsheetvariants
    (projectid, name, headerrow,
     sysnamecolumn, zsamplecolumn,
     systemcountcolumn, maxdistancecolumn)
VALUES ($1::int, $2::text, $3::int,
        $4::int, $5::int, $6::int, $7::int)
RETURNING id
`,

		// VSDS: update a variant's base fields
		"updatevariant": `
UPDATE vsds.spreadsheetvariants
SET name              = $3::text,
    headerrow         = $4::int,
    sysnamecolumn     = $5::int,
    zsamplecolumn     = $6::int,
    systemcountcolumn = $7::int,
    maxdistancecolumn = $8::int
WHERE id = $1::int AND projectid = $2::int
`,

		// VSDS: delete a variant (cascades to its checks)
		"deletevariant": `
DELETE FROM vsds.spreadsheetvariants
WHERE id = $1::int AND projectid = $2::int
`,

		// VSDS: insert a header check for a variant, returns new id
		"addvariantcheck": `
INSERT INTO vsds.spreadsheetvariant_checks
    (variantid, col, row, value)
VALUES ($1::int, $2::int, $3::int, $4::text)
RETURNING id
`,

		// VSDS: delete a header check by id within a variant
		"deletevariantcheck": `
DELETE FROM vsds.spreadsheetvariant_checks
WHERE id = $1::int AND variantid = $2::int
`,

		// admin: fetch id + customerid for a single cmdr
		"getcmdrbasic": `
SELECT id, customerid FROM common.cmdrs WHERE id = $1::int
`,

		// admin: list all commanders from view
		"listcmdrs": `
SELECT id, name, customerid, isowner, isadmin
FROM common.v_cmdrs
ORDER BY name
`,

		// admin: set isadmin on a commander, returns updated row
		"setcmdradmin": `
UPDATE common.cmdrs SET isadmin = $2::boolean
WHERE id = $1::int
RETURNING id, name, customerid, isowner,
          (isowner OR isadmin) AS isadmin
`,

		// self-delete: remove all personal data for a cmdr
		"deletecmdr": `SELECT common.deletecmdr($1::int)`,

		// bundles: list pending/queued bundles (vsds view)
		"listpendingbundles": `
SELECT * FROM bundles.v_vsds_bundles
WHERE status IN ('pending', 'queued')
ORDER BY id
`,

		// bundles: atomically claim a bundle for generation
		"setbundlegenerating": `
UPDATE bundles.bundles
SET status = 'generating'
WHERE id = $1 AND status IN ('pending', 'queued')
`,

		// bundles: mark a bundle as ready
		"setbundleready": `
UPDATE bundles.bundles
SET status = 'ready', generatedat = now(), errormessage = NULL
WHERE id = $1
`,

		// bundles: mark a bundle as failed
		"setbundleerror": `
UPDATE bundles.bundles
SET status = 'error', errormessage = $2
WHERE id = $1
`,

		// bundles: load VSDS-specific config for a bundle
		"getvsdsbundleconfig": `
SELECT subtype, allprojects
FROM bundles.vsds_bundles
WHERE bundleid = $1
`,

		// bundles: load project IDs for a VSDS bundle
		"getvsdsbundleprojects": `
SELECT projectid
FROM bundles.vsds_bundle_projects
WHERE bundleid = $1
ORDER BY projectid
`,

		// bundles: queue autoregen bundles for given project IDs
		"queueautoregen": `
SELECT bundles.queue_autoregen_bundles($1::int[])
`,

		// bundles: list all bundles (VSDS view)
		"listbundles": `
SELECT * FROM bundles.v_vsds_bundles ORDER BY id
`,

		// bundles: get a bundle by ID
		"getbundle": `
SELECT * FROM bundles.v_vsds_bundles WHERE id = $1
`,

		// bundles: create a VSDS bundle via DB function
		"createvsdsbundle": `
SELECT * FROM bundles.create_vsds_bundle(
    $1::varchar, $2::bool, $3::varchar,
    $4::bool, $5::int[]
)
`,

		// bundles: delete a bundle by ID
		"deletebundle": `
DELETE FROM bundles.bundles WHERE id = $1
`,

		// bundles: queue a bundle unless already generating
		"queuebundle": `
UPDATE bundles.bundles
SET status = 'queued'
WHERE id = $1 AND status != 'generating'
`,

		// bundles: check whether a bundle row exists
		"checkbundleexists": `
SELECT id FROM bundles.bundles WHERE id = $1
`,

		// bundles: update autoregen flag on a bundle
		"updatebundleautoregen": `
UPDATE bundles.bundles SET autoregen = $2 WHERE id = $1
`,

		// bundles: VSDS survey points — all projects
		"vsds_bundle_surveypts_all": `
SELECT sysname, zsample, x, y, z,
       corrected_n, maxdistance, rho, gc_x, gc_y, gc_z
FROM vsds.v_surveypoints
`,

		// bundles: VSDS survey points — filtered by project
		"vsds_bundle_surveypts_proj": `
SELECT sysname, zsample, x, y, z,
       corrected_n, maxdistance, rho, gc_x, gc_y, gc_z
FROM vsds.v_surveypoints
WHERE surveyid IN (
    SELECT id FROM vsds.surveys
    WHERE projectid = ANY($1::int[])
)
`,

		// bundles: VSDS surveys — all projects
		"vsds_bundle_surveys_all": `
SELECT projectname, rho_max, x, z,
       column_dev, gc_x, gc_z, points
FROM vsds.v_surveys
`,

		// bundles: VSDS surveys — filtered by project
		"vsds_bundle_surveys_proj": `
SELECT projectname, rho_max, x, z,
       column_dev, gc_x, gc_z, points
FROM vsds.v_surveys
WHERE projectid = ANY($1::int[])
`,

		// VSDS sectors: aggregate survey point density into voxels
		"vsds_sectors": `
SELECT gc_x, gc_z, y_min, y_max,
       rho_min, rho_avg, rho_max, cnt
FROM vsds.sectors($1::float8, $2::float8)
`,

		// refresh both survey matviews via SECURITY DEFINER
		// function; avoids requiring edservice to own the views
		"refreshsurveymatviews": `
SELECT vsds.refresh_survey_matviews()
`,
	}


	logger log.Logger
)

type DBPool struct {
	ctx context.Context
	pool *pgxpool.Pool
}

// init the DBPool and store it in the global variable
func Init(cfg *config.DBConfig) error {
	var err error

	logger = log.GetLogger("db")

	sslmode := "disable"
	if cfg.SSL {
		sslmode = "prefer"
	}
	dbcfg, err := pgxpool.ParseConfig("sslmode=" + sslmode)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to parse config")
		return err
	}
	dbcfg.MaxConnLifetime = 8 * time.Hour
	dbcfg.MaxConns = cfg.MaxConns
	dbcfg.MinConns = cfg.MinConns
	dbcfg.AfterConnect = afterConn
	dbcfg.ConnConfig.Host = cfg.Host
	dbcfg.ConnConfig.Port = 5432
	dbcfg.ConnConfig.Database = cfg.Database
	dbcfg.ConnConfig.User = cfg.User
	dbcfg.ConnConfig.Password = cfg.Password
	if cfg.Port != nil {
		dbcfg.ConnConfig.Port = (*cfg.Port)
	}

	dbp := DBPool{
		ctx: context.Background(),
	}

	if dbp.pool, err = pgxpool.NewWithConfig(dbp.ctx, dbcfg); err != nil {
		logger.Error().Err(err).Msg("Unable to initialize pool")
		return err
	}

	conn, err := dbp.pool.Acquire(dbp.ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to acquire connection from pool")
		return err
	}
	defer conn.Release()

	Pool = &dbp
	logger.Info().Msg("Database subsystem started")
	return nil
}

func afterConn(ctx context.Context, dbc *pgx.Conn) error {
	for name, query := range prepared {
		if _, err := dbc.Prepare(ctx, name, query); err != nil {
			logger.Error().Err(err).Str("qname", name).Str("query", query).
				Msg("Error while preparing query")
			return errors.Join(err, fmt.Errorf("Error while preparing %s", name))
		}
	}
	return nil
}

func (p *DBPool) Close() error {
	p.pool.Close()
	return nil
}

// Transaction holds a long-running transaction with checkpoint support.
// Each operation saves a savepoint on success; on failure it rolls back
// to the last savepoint, keeping the transaction alive.
type Transaction struct {
	ctx           context.Context
	conn          *pgxpool.Conn
	tx            pgx.Tx
	checkpoint    string
	checkpointIdx int
}

// StartLongTxn acquires a dedicated connection, opens a RepeatableRead
// transaction, and sets an initial checkpoint. The caller must call
// Close() when done.
func (p *DBPool) StartLongTxn() (*Transaction, error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return nil, err
	}

	tx, err := conn.BeginTx(p.ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		conn.Release()
		logger.Error().Err(err).Caller().
			Msg("Error opening long transaction")
		return nil, err
	}

	dbt := &Transaction{
		ctx:  p.ctx,
		conn: conn,
		tx:   tx,
	}

	if err := dbt.saveCheckpoint(); err != nil {
		tx.Rollback(p.ctx)
		conn.Release()
		return nil, err
	}

	return dbt, nil
}

// Close commits the transaction (io.Closer semantics).
// On commit failure it rolls back to the last checkpoint and
// commits that, preserving successfully checkpointed work.
func (t *Transaction) Close() error {
	defer t.conn.Release()

	logger.Info().Msg("Closing transaction...")
	if err := t.tx.Commit(t.ctx); err != nil {
		logger.Error().Err(err).Caller().
			Msg("Commit failed, rolling back to last checkpoint")
		if rerr := t.rollbackToCheckpoint(); rerr != nil {
			t.tx.Rollback(t.ctx)
			return rerr
		}
		if cerr := t.tx.Commit(t.ctx); cerr != nil {
			t.tx.Rollback(t.ctx)
			logger.Error().Err(cerr).Caller().
				Msg("Commit after checkpoint rollback failed")
			return cerr
		}
	}

	return nil
}

// saveCheckpoint creates a new savepoint and records it as the
// current checkpoint.
func (t *Transaction) saveCheckpoint() error {
	t.checkpointIdx++
	name := fmt.Sprintf("sp_%d", t.checkpointIdx)
	if _, err := t.tx.Exec(t.ctx, "SAVEPOINT "+name); err != nil {
		logger.Error().Err(err).Caller().
			Str("savepoint", name).
			Msg("Error creating savepoint")
		return err
	}
	t.checkpoint = name
	return nil
}

// rollbackToCheckpoint rolls back to the last saved savepoint without
// aborting the transaction, leaving it open for further operations.
func (t *Transaction) rollbackToCheckpoint() error {
	if t.checkpoint == "" {
		return nil
	}
	if _, err := t.tx.Exec(
		t.ctx, "ROLLBACK TO SAVEPOINT "+t.checkpoint,
	); err != nil {
		logger.Error().Err(err).Caller().
			Str("savepoint", t.checkpoint).
			Msg("Error rolling back to savepoint")
		return err
	}
	return nil
}
