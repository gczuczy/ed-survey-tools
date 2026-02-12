package db

import (
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID int `db:"id",json:"id"`
	Name string `db:"name",json:"name"`
	CustomerID int64 `db:"customerid",json:"customerid"`
	IsOwner bool `db:"isowner",json:"isowner"`
	IsAdmin bool `db:"isadmin",json:"isadmin"`
}

func (p *DBPool) LoginCMDR(name string, customerid int64) (*User, error) {
	var rows pgx.Rows
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	if rows, err = conn.Query(p.ctx, "logincmdr", name, customerid);  err != nil {
		return nil, err
	}
	defer rows.Close()

	user, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, err
	}

	return &user[0], nil
}
