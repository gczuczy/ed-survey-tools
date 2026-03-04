package db

import (
	"fmt"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/gczuczy/ed-survey-tools/pkg/vsds"
)

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
	ID int64 `db:"id"`
	Name string `db:"name"`
	X float32 `db:"x"`
	Y float32 `db:"y"`
	Z float32 `db:"z"`
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
			err = ErrDuplicate
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
			err = ErrDuplicate
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

func (p *DBPool) AddSurvey(m *vsds.Survey) (err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return err
	}
	defer conn.Release()
	tx, err := conn.Begin(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Error while opening txn")
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(p.ctx)
			return
		}
		if tx.Commit(p.ctx) != nil {
			tx.Rollback(p.ctx)
		}
	}()

	var rows pgx.Rows

	if rows, err = tx.Query(p.ctx, "addsheetsurvey",	m.CMDR, m.Project);  err != nil {
		logger.Error().Err(err).Caller().Str("query", "addsheetsurvey").
			Msg("Error while performing query")
		return err
	}

	if !rows.Next() {
		logger.Error().Caller().Msg("No surveyid returned")
		return fmt.Errorf("No surveyid returned")
	}

	var vs []any
	if vs, err = rows.Values(); err != nil {
		err := errors.Join(err, fmt.Errorf("Fuck golang's error handling"))
		logger.Error().Err(err).Caller().Msg("Golang's error handling is definitely awesome")
		return err
	}

	mid, ok := vs[0].(int32)
	if !ok {
		rows.Close()
		err := errors.Join(err, fmt.Errorf("Fuck golang's error handling again, %v/%T -> %v", vs[0], vs[0], mid))
		logger.Error().Err(err).Caller().Msgf("Type mismatch %v/%T -> %v",
			vs[0], vs[0], mid)
		return err
	}
	rows.Close()

	for _, dp := range m.SurveyPoints {
		if rows, err = tx.Query(p.ctx, "setsystem", dp.EDSMID, dp.SystemName,
			dp.X, dp.Y, dp.Z);  err != nil {
			err := errors.Join(err, fmt.Errorf("Query(setsystem) error"))
			logger.Error().Err(err).Caller().Str("query", "setsystem").
				Msg("Query error")
			return err
		}

		sys, err := pgx.CollectRows(rows, pgx.RowToStructByName[System])
		if err != nil {
			err = errors.Join(err, fmt.Errorf("Unable to decode data %+v", rows))
			logger.Error().Err(err).Caller().Interface("data", rows).
				Msgf("Unable to decode data")
			return err
		}
		rows.Close()
		if _, err = tx.Exec(p.ctx, "addsurveypoint", mid, sys[0].ID, dp.ZSample,
			dp.Count, dp.MaxDistance); err != nil {
			err = errors.Join(err, fmt.Errorf("Error while inserting surveypoint"))
			logger.Error().Err(err).Caller().Str("query", "addsurveypoint").
				Msg("Query error")
			return err
		}
	}

	return nil
}
