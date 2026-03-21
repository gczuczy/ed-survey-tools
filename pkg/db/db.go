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

	ErrNotFound     = fmt.Errorf("not found")
	ErrDuplicate    = fmt.Errorf("already exists")
	ErrAlreadyQueued = fmt.Errorf("processing already queued or in progress")

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

		// record a sheet processing outcome
		"recordsheetresult": `
INSERT INTO vsds.sheet_processing (procid, sheetid, success, message)
VALUES ($1::int, $2::int, $3::boolean, NULLIF($4::text, ''))
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

	dbcfg, err := pgxpool.ParseConfig("")
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
