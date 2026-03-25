package db

import (
	"encoding/gob"
	"errors"

	"github.com/jackc/pgx/v5"
)

func init() {
	gob.Register(&User{})
}

type User struct {
	ID int `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	CustomerID *int64 `db:"customerid" json:"customerid"`
	IsOwner bool `db:"isowner" json:"isowner"`
	IsAdmin bool `db:"isadmin" json:"isadmin"`
}

func (p *DBPool) ListCMDRs() ([]*User, error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(p.ctx, "listcmdrs")
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "listcmdrs").
			Msg("Error while executing query")
		return nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[User])
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return nil, err
	}
	return users, nil
}

func (p *DBPool) SetCMDRAdmin(id int, isAdmin bool) (*User, error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return nil, err
	}
	defer conn.Release()

	// When granting admin, verify the cmdr exists and has a customerid.
	if isAdmin {
		var cmdrID int
		var customerID *int64
		err = conn.QueryRow(p.ctx, "getcmdrbasic", id).
			Scan(&cmdrID, &customerID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNotFound
			}
			logger.Error().Err(err).Caller().
				Str("query", "getcmdrbasic").
				Msg("Error while checking cmdr")
			return nil, err
		}
		if customerID == nil {
			return nil, ErrCustomerIDRequired
		}
	}

	rows, err := conn.Query(p.ctx, "setcmdradmin", id, isAdmin)
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "setcmdradmin").
			Msg("Error while executing query")
		return nil, err
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[User])
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return nil, err
	}
	if len(users) == 0 {
		return nil, ErrNotFound
	}
	return users[0], nil
}

func (p *DBPool) NullifyCMDRCustomerID(id int) error {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Unable to acquire connection from pool")
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(p.ctx, "deletecmdr", id)
	if err != nil {
		logger.Error().Err(err).Caller().
			Str("query", "deletecmdr").
			Msg("Error while executing query")
		return err
	}
	return nil
}

func (p *DBPool) LoginCMDR(name string, customerid int64) (*User, error) {
	var rows pgx.Rows
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		logger.Error().Err(err).Caller().Msg("Unable to acquire connection from pool")
		return nil, err
	}
	defer conn.Release()

	if rows, err = conn.Query(p.ctx, "logincmdr", name, customerid);  err != nil {
		logger.Error().Err(err).Caller().Str("query", "logincmdr").
			Msg("Error while executing query")
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		logger.Error().Err(err).Caller().
			Msg("Error while reading results")
		return nil, err
	}

	return &user[0], nil
}
