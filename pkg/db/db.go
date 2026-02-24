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

	prepared = map[string]string{
		// logincmdr
		"logincmdr": `
SELECT * FROM common.logincmdr($1::text, $2::bigint)
`,

		// add a sheet survey
		"addsheetsurvey": `
SELECT density.addsheetsurvey($1::text, $2::text)
`,
		"setsystem": `
INSERT INTO common.systems (edsmid, name, x, y, z)
VALUES ($1::bigint, $2::text, $3::float, $4::float, $5::float)
ON CONFLICT (edsmid) DO UPDATE SET edsmid = EXCLUDED.edsmid
RETURNING *
`,
		// surveyid, sysname, x,y,z, syscount, maxdistance
		"addsurveypoint": `
INSERT INTO density.surveypoints (surveyid, sysid, zsample, syscount, maxdistance)
VALUES ($1::int, $2::bigint, $3::int, $4::int, $5::real)
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
