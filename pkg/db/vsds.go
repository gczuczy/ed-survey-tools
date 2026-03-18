package db

import (
	"fmt"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	vsdstypes "github.com/gczuczy/ed-survey-tools/pkg/vsds/types"
)

// FolderProcessingSummary holds aggregate statistics for one
// folder extraction run, returned by GetFolderProcessingDetails.
type FolderProcessingSummary struct {
	ReceivedAt      time.Time  `db:"receivedat"        json:"received_at"`
	StartedAt       *time.Time `db:"startedat"         json:"started_at,omitempty"`
	FinishedAt      *time.Time `db:"finishedat"        json:"finished_at,omitempty"`
	DocumentsTotal  int64      `db:"documents_total"   json:"documents_total"`
	SheetsTotal     int64      `db:"sheets_total"      json:"sheets_total"`
	SheetsSuccess   int64      `db:"sheets_success"    json:"sheets_success"`
	SheetsFailed    int64      `db:"sheets_failed"     json:"sheets_failed"`
	SurveysIngested int64      `db:"surveys_ingested"  json:"surveys_ingested"`
	PointsIngested  int64      `db:"points_ingested"   json:"points_ingested"`
	CmdrsCount      int64      `db:"cmdrs_count"       json:"cmdrs_count"`
}

// FolderProcessingSheetRow is one failing sheet row returned by
// GetFolderProcessingDetails, grouped into documents by the caller.
type FolderProcessingSheetRow struct {
	DocID       int    `db:"doc_id"`
	DocGCPID    string `db:"gcpid"`
	DocName     string `db:"doc_name"`
	ContentType string `db:"contenttype"`
	SheetID     int    `db:"sheet_id"`
	SheetName   string `db:"sheet_name"`
	Message     string `db:"message"`
}

type VSDSFolder struct {
	FolderID   int        `db:"folderid" json:"id"`
	Name       string     `db:"name" json:"name"`
	GCPID      string     `db:"gcpid" json:"gcpid"`
	ReceivedAt *time.Time `db:"receivedat" json:"received_at,omitempty"`
	StartedAt  *time.Time `db:"startedat" json:"started_at,omitempty"`
	FinishedAt *time.Time `db:"finishedat" json:"finished_at,omitempty"`
	InProgress *int64     `db:"inprogress" json:"in_progress,omitempty"`
	Finished   *int64     `db:"finished" json:"finished,omitempty"`
	Failed     *int64     `db:"failed" json:"failed,omitempty"`
}

type System struct {
	ID     int64   `db:"id"`
	EDSMID int64   `db:"edsmid"`
	Name   string  `db:"name"`
	X      float32 `db:"x"`
	Y      float32 `db:"y"`
	Z      float32 `db:"z"`
}

type VSDSProject struct {
	ID       int    `db:"id"       json:"id"`
	Name     string `db:"name"     json:"name"`
	ZSamples []int  `db:"zsamples" json:"zsamples"`
}

func (p *DBPool) ListProjects() (projects []VSDSProject, err error) {
	var (
		rows pgx.Rows
	)

	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	if rows, err = conn.Query(p.ctx, "listprojects");  err != nil {
		logger.Error().Err(err).Caller().Str("query", "listprojects").
			Msg("Error while executing query")
		return
	}
	defer rows.Close()

	projects, err = pgx.CollectRows(rows, pgx.RowToStructByName[VSDSProject])
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return
	}

	return projects, nil
}

func (p *DBPool) AddProject(name string) (project VSDSProject, err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	rows, err := conn.Query(p.ctx, "addproject", name)
	if err != nil {
		logger.Error().Err(err).Caller().Str("query", "addproject").
			Msg("Error while executing query")
		return
	}
	defer rows.Close()

	projects, err := pgx.CollectRows(rows, pgx.RowToStructByName[VSDSProject])
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return
	}

	if len(projects) == 0 {
		err = fmt.Errorf("No project returned after insert")
		return
	}

	return projects[0], nil
}

func (p *DBPool) GetProject(id int) (project VSDSProject, err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	rows, err := conn.Query(p.ctx, "getproject", id)
	if err != nil {
		logger.Error().Err(err).Caller().Str("query", "getproject").
			Msg("Error while executing query")
		return
	}
	defer rows.Close()

	projects, err := pgx.CollectRows(rows, pgx.RowToStructByName[VSDSProject])
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Error while reading results")
		return
	}

	if len(projects) == 0 {
		err = ErrNotFound
		return
	}

	return projects[0], nil
}

func (p *DBPool) SetZSamples(projectID int, zsamples []int) (project VSDSProject, err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Error while opening txn")
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback(p.ctx)
			return
		}
		if cerr := tx.Commit(p.ctx); cerr != nil {
			tx.Rollback(p.ctx)
			err = cerr
		}
	}()

	var rows pgx.Rows
	if rows, err = tx.Query(p.ctx, "getproject", projectID); err != nil {
		logger.Error().Err(err).Caller().Str("query", "getproject").
			Msg("Error while executing query")
		return
	}
	existing, cerr := pgx.CollectRows(rows, pgx.RowToStructByName[VSDSProject])
	if cerr != nil {
		err = cerr
		logger.Error().Err(err).Caller().Msg("Error while reading results")
		return
	}
	if len(existing) == 0 {
		err = ErrNotFound
		return
	}

	if _, err = tx.Exec(p.ctx, "deleteprojectzsamples", projectID); err != nil {
		logger.Error().Err(err).Caller().Str("query", "deleteprojectzsamples").
			Msg("Error while executing query")
		return
	}

	for _, zsample := range zsamples {
		if _, err = tx.Exec(p.ctx, "insertprojectzsample", projectID, zsample); err != nil {
			logger.Error().Err(err).Caller().Str("query", "insertprojectzsample").
				Int("zsample", zsample).Msg("Error while executing query")
			return
		}
	}

	if rows, err = tx.Query(p.ctx, "getproject", projectID); err != nil {
		logger.Error().Err(err).Caller().Str("query", "getproject").
			Msg("Error while executing query")
		return
	}
	updated, cerr := pgx.CollectRows(rows, pgx.RowToStructByName[VSDSProject])
	if cerr != nil {
		err = cerr
		logger.Error().Err(err).Caller().Msg("Error while reading results")
		return
	}
	if len(updated) == 0 {
		err = ErrNotFound
		return
	}

	project = updated[0]
	return
}

func (p *DBPool) AddZSample(projectID, zsample int) (project VSDSProject, err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Error while opening txn")
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback(p.ctx)
			return
		}
		if cerr := tx.Commit(p.ctx); cerr != nil {
			tx.Rollback(p.ctx)
			err = cerr
		}
	}()

	var rows pgx.Rows
	if rows, err = tx.Query(p.ctx, "getproject", projectID); err != nil {
		logger.Error().Err(err).Caller().Str("query", "getproject").
			Msg("Error while executing query")
		return
	}
	existing, cerr := pgx.CollectRows(rows, pgx.RowToStructByName[VSDSProject])
	if cerr != nil {
		err = cerr
		logger.Error().Err(err).Caller().Msg("Error while reading results")
		return
	}
	if len(existing) == 0 {
		err = ErrNotFound
		return
	}

	if _, err = tx.Exec(p.ctx, "insertprojectzsample", projectID, zsample); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			err = newQueryError(ErrDuplicate, map[string]any{
				"projectid": projectID,
				"zsample":   zsample,
			})
		} else {
			logger.Error().Err(err).Caller().Str("query", "insertprojectzsample").
				Msg("Error while executing query")
		}
		return
	}

	if rows, err = tx.Query(p.ctx, "getproject", projectID); err != nil {
		logger.Error().Err(err).Caller().Str("query", "getproject").
			Msg("Error while executing query")
		return
	}
	updated, cerr := pgx.CollectRows(rows, pgx.RowToStructByName[VSDSProject])
	if cerr != nil {
		err = cerr
		logger.Error().Err(err).Caller().Msg("Error while reading results")
		return
	}
	if len(updated) == 0 {
		err = ErrNotFound
		return
	}

	project = updated[0]
	return
}

func (p *DBPool) DeleteZSample(projectID, zsample int) (project VSDSProject, err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Error while opening txn")
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback(p.ctx)
			return
		}
		if cerr := tx.Commit(p.ctx); cerr != nil {
			tx.Rollback(p.ctx)
			err = cerr
		}
	}()

	var rows pgx.Rows
	if rows, err = tx.Query(p.ctx, "getproject", projectID); err != nil {
		logger.Error().Err(err).Caller().Str("query", "getproject").
			Msg("Error while executing query")
		return
	}
	existing, cerr := pgx.CollectRows(rows, pgx.RowToStructByName[VSDSProject])
	if cerr != nil {
		err = cerr
		logger.Error().Err(err).Caller().Msg("Error while reading results")
		return
	}
	if len(existing) == 0 {
		err = ErrNotFound
		return
	}

	tag, execErr := tx.Exec(p.ctx, "deleteprojectzsample", projectID, zsample)
	if execErr != nil {
		err = execErr
		logger.Error().Err(err).Caller().Str("query", "deleteprojectzsample").
			Msg("Error while executing query")
		return
	}
	if tag.RowsAffected() == 0 {
		err = ErrNotFound
		return
	}

	if rows, err = tx.Query(p.ctx, "getproject", projectID); err != nil {
		logger.Error().Err(err).Caller().Str("query", "getproject").
			Msg("Error while executing query")
		return
	}
	updated, cerr := pgx.CollectRows(rows, pgx.RowToStructByName[VSDSProject])
	if cerr != nil {
		err = cerr
		logger.Error().Err(err).Caller().Msg("Error while reading results")
		return
	}
	if len(updated) == 0 {
		err = ErrNotFound
		return
	}

	project = updated[0]
	return
}

func (p *DBPool) ListFolders() (folders []VSDSFolder, err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	rows, err := conn.Query(p.ctx, "listfolders")
	if err != nil {
		logger.Error().Err(err).Caller().Str("query", "listfolders").
			Msg("Error while executing query")
		return
	}
	defer rows.Close()

	folders, err = pgx.CollectRows(rows, pgx.RowToStructByName[VSDSFolder])
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return
	}

	return folders, nil
}

func (p *DBPool) AddFolder(gcpid, name string) (folder VSDSFolder, err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Error while opening txn")
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback(p.ctx)
			return
		}
		if cerr := tx.Commit(p.ctx); cerr != nil {
			tx.Rollback(p.ctx)
			err = cerr
		}
	}()

	var id int
	if err = tx.QueryRow(p.ctx, "addfolder", gcpid, name).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			err = newQueryError(ErrDuplicate, map[string]any{
				"gcpid": gcpid,
				"name":  name,
			})
		} else {
			logger.Error().Err(err).Caller().Str("query", "addfolder").
				Msg("Error while executing query")
		}
		return
	}

	rows, err := tx.Query(p.ctx, "getfolder", id)
	if err != nil {
		logger.Error().Err(err).Caller().Str("query", "getfolder").
			Msg("Error while executing query")
		return
	}
	defer rows.Close()

	folders, cerr := pgx.CollectRows(rows, pgx.RowToStructByName[VSDSFolder])
	if cerr != nil {
		err = cerr
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return
	}

	if len(folders) == 0 {
		err = fmt.Errorf("No folder returned after insert")
		return
	}

	folder = folders[0]
	return
}

func (p *DBPool) QueueFolderProcessing(folderID int) (err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	tx, err := conn.BeginTx(p.ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Error while opening txn")
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback(p.ctx)
			return
		}
		if cerr := tx.Commit(p.ctx); cerr != nil {
			tx.Rollback(p.ctx)
			err = cerr
		}
	}()

	// Verify folder exists
	var rows pgx.Rows
	if rows, err = tx.Query(p.ctx, "getfolder", folderID); err != nil {
		logger.Error().Err(err).Caller().Str("query", "getfolder").
			Msg("Error while executing query")
		return
	}
	folders, cerr := pgx.CollectRows(rows, pgx.RowToStructByName[VSDSFolder])
	if cerr != nil {
		err = cerr
		logger.Error().Err(err).Caller().Msg("Error while reading results")
		return
	}
	if len(folders) == 0 {
		err = ErrNotFound
		return
	}

	// Check for active or pending processing within the transaction
	var hasActive bool
	if err = tx.QueryRow(p.ctx, "checkfolderprocessing", folderID).Scan(&hasActive); err != nil {
		logger.Error().Err(err).Caller().Str("query", "checkfolderprocessing").
			Msg("Error while executing query")
		return
	}
	if hasActive {
		err = ErrAlreadyQueued
		return
	}

	// Queue the processing request
	var newID int
	if err = tx.QueryRow(p.ctx, "insertfolderprocessing", folderID).Scan(&newID); err != nil {
		logger.Error().Err(err).Caller().Str("query", "insertfolderprocessing").
			Msg("Error while executing query")
		return
	}

	return nil
}

func (p *DBPool) FetchPendingFolderProcessing() (job *vsdstypes.FolderProcessingJob, err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	tx, err := conn.BeginTx(p.ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Error while opening txn")
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback(p.ctx)
			return
		}
		if cerr := tx.Commit(p.ctx); cerr != nil {
			tx.Rollback(p.ctx)
			err = cerr
		}
	}()

	var procID, folderID int
	var gcpID string
	err = tx.QueryRow(p.ctx, "fetchpendingfolderprocessing").Scan(&procID, &folderID, &gcpID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
		return nil, nil
	}
	if err != nil {
		logger.Error().Err(err).Caller().Str("query", "fetchpendingfolderprocessing").
			Msg("Error while executing query")
		return
	}

	if _, err = tx.Exec(p.ctx, "startfolderprocessing", procID); err != nil {
		logger.Error().Err(err).Caller().Str("query", "startfolderprocessing").
			Msg("Error while executing query")
		return
	}

	job = &vsdstypes.FolderProcessingJob{
		ProcID:   procID,
		FolderID: folderID,
		GCPID:    gcpID,
	}
	return
}

func (p *DBPool) DeleteFolder(id int) (err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	tag, err := conn.Exec(p.ctx, "deletefolder", id)
	if err != nil {
		logger.Error().Err(err).Caller().Str("query", "deletefolder").
			Msg("Error while executing query")
		return
	}

	if tag.RowsAffected() == 0 {
		err = ErrNotFound
		return
	}

	return nil
}

func (p *DBPool) FinishFolderProcessing(procID int) (err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	tag, err := conn.Exec(p.ctx, "finishfolderprocessing", procID)
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "finishfolderprocessing").
			Msg("Error while executing query")
		return
	}

	if tag.RowsAffected() == 0 {
		err = ErrNotFound
	}
	return
}

// GetFolderProcessingDetails returns the folder name, summary
// statistics, and failing sheet rows for the most recent processing
// run of the given folder. Returns ErrNotFound when no run exists.
func (p *DBPool) GetFolderProcessingDetails(folderID int) (
	folderName string,
	summary FolderProcessingSummary,
	sheetRows []FolderProcessingSheetRow,
	err error,
) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return
	}
	defer conn.Release()

	var procID int
	err = conn.QueryRow(
		p.ctx, "getlastfolderprocessing", folderID,
	).Scan(&procID, &folderName)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrNotFound
		return
	}
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "getlastfolderprocessing").
			Msg("Error while executing query")
		return
	}

	var rows pgx.Rows
	if rows, err = conn.Query(
		p.ctx, "getfolderprocessingsummary", procID,
	); err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "getfolderprocessingsummary").
			Msg("Error while executing query")
		return
	}
	defer rows.Close()

	summaries, cerr := pgx.CollectRows(
		rows, pgx.RowToStructByName[FolderProcessingSummary],
	)
	if cerr != nil {
		err = cerr
		logger.Error().Err(err).Caller().
			Msg("Error while reading summary results")
		return
	}
	if len(summaries) == 0 {
		err = ErrNotFound
		return
	}
	summary = summaries[0]

	var srows pgx.Rows
	if srows, err = conn.Query(
		p.ctx, "getfolderprocessingsheets", folderID, procID,
	); err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "getfolderprocessingsheets").
			Msg("Error while executing query")
		return
	}
	defer srows.Close()

	sheetRows, err = pgx.CollectRows(
		srows, pgx.RowToStructByName[FolderProcessingSheetRow],
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading sheet results")
		return
	}

	return
}

// DeleteFolderSpreadsheets removes all spreadsheets belonging to the
// given folder, cascading into surveys and surveypoints via the DB
// foreign key constraints. Rolls back to the last checkpoint on error.
func (t *Transaction) DeleteFolderSpreadsheets(folderID int) error {
	_, err := t.tx.Exec(
		t.ctx, "deletefolderspreadsheets", folderID,
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "deletefolderspreadsheets").
			Int("folderid", folderID).
			Msg("Error while executing query")
		if rerr := t.rollbackToCheckpoint(); rerr != nil {
			logger.Error().Err(rerr).Caller().
				Msg("Error rolling back to checkpoint")
		}
		return newQueryError(err, map[string]any{
			"folderid": folderID,
		})
	}

	return t.saveCheckpoint()
}

// Savepoint creates a new savepoint, recording it as the current
// checkpoint. Callers use this to protect blocks of work that span
// multiple Transaction method calls (e.g. system lookups followed by
// survey insertion).
func (t *Transaction) Savepoint() error {
	return t.saveCheckpoint()
}

// Rollback rolls back to the last savepoint without aborting the
// transaction, leaving it valid for further operations.
func (t *Transaction) Rollback() error {
	return t.rollbackToCheckpoint()
}

// AddSpreadsheet inserts a spreadsheet file record within the long
// transaction and returns the new row id.
func (t *Transaction) AddSpreadsheet(
	folderID int,
	gcpid, name, contentType string,
) (int, error) {
	var id int
	err := t.tx.QueryRow(
		t.ctx, "addspreadsheet",
		folderID, gcpid, name, contentType,
	).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "addspreadsheet").
			Str("gcpid", gcpid).
			Msg("Error while executing query")
		if rerr := t.rollbackToCheckpoint(); rerr != nil {
			logger.Error().Err(rerr).Caller().
				Msg("Error rolling back to checkpoint")
		}
		return 0, newQueryError(err, map[string]any{
			"contenttype": contentType,
			"folderid":    folderID,
			"gcpid":       gcpid,
			"name":        name,
		})
	}
	return id, t.saveCheckpoint()
}

// AddSheet inserts a sheet record within the long transaction and
// returns the new row id. Pass name=nil for implicit CSV sheets.
func (t *Transaction) AddSheet(
	spreadsheetID int,
	name *string,
) (int, error) {
	var id int
	err := t.tx.QueryRow(
		t.ctx, "addsheet", spreadsheetID, name,
	).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "addsheet").
			Msg("Error while executing query")
		if rerr := t.rollbackToCheckpoint(); rerr != nil {
			logger.Error().Err(rerr).Caller().
				Msg("Error rolling back to checkpoint")
		}
		var nameParam any
		if name != nil {
			nameParam = *name
		}
		return 0, newQueryError(err, map[string]any{
			"name":          nameParam,
			"spreadsheetid": spreadsheetID,
		})
	}
	return id, t.saveCheckpoint()
}

// UpsertCmdr inserts or retrieves a CMDR by name within the long
// transaction and returns the row id.
func (t *Transaction) UpsertCmdr(name string) (int, error) {
	var id int
	err := t.tx.QueryRow(
		t.ctx, "upsertcmdr", name,
	).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "upsertcmdr").
			Str("name", name).
			Msg("Error while executing query")
		if rerr := t.rollbackToCheckpoint(); rerr != nil {
			logger.Error().Err(rerr).Caller().
				Msg("Error rolling back to checkpoint")
		}
		return 0, newQueryError(err, map[string]any{
			"name": name,
		})
	}
	return id, t.saveCheckpoint()
}

// LookupProject resolves a project id by name within the long
// transaction. Returns ErrNotFound when no matching project exists;
// only rolls back to the last checkpoint on a real SQL error.
func (t *Transaction) LookupProject(name string) (int, error) {
	var id int
	err := t.tx.QueryRow(
		t.ctx, "lookupproject", name,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "lookupproject").
			Str("name", name).
			Msg("Error while executing query")
		if rerr := t.rollbackToCheckpoint(); rerr != nil {
			logger.Error().Err(rerr).Caller().
				Msg("Error rolling back to checkpoint")
		}
		return 0, newQueryError(err, map[string]any{
			"name": name,
		})
	}
	return id, nil
}

// AddSurvey inserts a survey and all its points within the long
// transaction. systems maps each SurveyPoint.SystemName to its
// resolved db.System row.
func (t *Transaction) AddSurvey(
	sheetID, projectID, cmdrID int,
	points []vsdstypes.SurveyPoint,
	systems map[string]System,
) error {
	var surveyID int
	err := t.tx.QueryRow(
		t.ctx, "addsurvey",
		projectID, cmdrID, sheetID,
	).Scan(&surveyID)
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "addsurvey").
			Msg("Error while executing query")
		if rerr := t.rollbackToCheckpoint(); rerr != nil {
			logger.Error().Err(rerr).Caller().
				Msg("Error rolling back to checkpoint")
		}
		return newQueryError(err, map[string]any{
			"cmdrid":    cmdrID,
			"projectid": projectID,
			"sheetid":   sheetID,
		})
	}

	for _, sp := range points {
		sys, ok := systems[sp.SystemName]
		if !ok {
			// System was not resolved during lookup;
			// the caller has already noted the error.
			continue
		}
		_, err = t.tx.Exec(
			t.ctx, "addsurveypoint",
			surveyID, sys.ID,
			sp.ZSample, sp.Count, sp.MaxDistance,
		)
		if err != nil {
			logger.Error().Err(err).Caller().
				Str("query", "addsurveypoint").
				Str("system", sp.SystemName).
				Int("zsample", sp.ZSample).
				Int("count", sp.Count).
				Float32("maxdistance", sp.MaxDistance).
				Msg("Error while executing query")
			if rerr := t.rollbackToCheckpoint(); rerr != nil {
				logger.Error().Err(rerr).Caller().
					Msg("Error rolling back to checkpoint")
			}
			return newQueryError(err, map[string]any{
				"count":       sp.Count,
				"maxdistance": sp.MaxDistance,
				"surveyid":    surveyID,
				"system":      sp.SystemName,
				"zsample":     sp.ZSample,
			})
		}
	}

	return t.saveCheckpoint()
}

// RecordSheetResult inserts a sheet_processing row recording the
// outcome of processing one sheet. An empty message is stored as
// NULL via the NULLIF in the prepared query.
func (t *Transaction) RecordSheetResult(
	procID, sheetID int,
	success bool,
	message string,
) error {
	_, err := t.tx.Exec(
		t.ctx, "recordsheetresult",
		procID, sheetID, success, message,
	)
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "recordsheetresult").
			Msg("Error while executing query")
		if rerr := t.rollbackToCheckpoint(); rerr != nil {
			logger.Error().Err(rerr).Caller().
				Msg("Error rolling back to checkpoint")
		}
		return newQueryError(err, map[string]any{
			"procid":  procID,
			"sheetid": sheetID,
			"success": success,
		})
	}
	return t.saveCheckpoint()
}
