package db

import (
	"fmt"
	"errors"

	"github.com/jackc/pgx/v5"
	ds "github.com/gczuczy/ed-survey-tools/pkg/densitysurvey"
)

type System struct {
	ID int64 `db:"id"`
	Name string `db:"name"`
	X float32 `db:"x"`
	Y float32 `db:"y"`
	Z float32 `db:"z"`
}

func (p *DBPool) AddSurvey(m *ds.Survey) (err error) {
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
