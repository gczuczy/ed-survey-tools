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
		return err
	}
	defer conn.Release()
	tx, err := conn.Begin(p.ctx)
	if err != nil {
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
		return err
	}

	if !rows.Next() {
		return fmt.Errorf("No surveyid returned")
	}

	var vs []any
	if vs, err = rows.Values(); err != nil {
		return errors.Join(err, fmt.Errorf("Fuck golang's error handling"))
	}

	mid, ok := vs[0].(int32)
	if !ok {
		rows.Close()
		return errors.Join(err, fmt.Errorf("Fuck golang's error handling again, %v/%T -> %v", vs[0], vs[0], mid))
	}
	rows.Close()

	for _, dp := range m.SurveyPoints {
		if rows, err = tx.Query(p.ctx, "setsystem", dp.EDSMID, dp.SystemName,
			dp.X, dp.Y, dp.Z);  err != nil {
			return errors.Join(err, fmt.Errorf("Query(setsystem) error"))
		}

		sys, err := pgx.CollectRows(rows, pgx.RowToStructByName[System])
		if err != nil {
			return errors.Join(err, fmt.Errorf("Unable to decode data %+v", rows))
		}
		rows.Close()
		if _, err = tx.Exec(p.ctx, "addsurveypoint", mid, sys[0].ID, dp.ZSample,
			dp.Count, dp.MaxDistance); err != nil {
			return errors.Join(err, fmt.Errorf("Error while inserting surveypoint"))
		}
	}

	return nil
}
