package dbUtil

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type RowHandler func(*sqlx.Rows) error

func QuickRows(db sqlx.Queryer, qry squirrel.SelectBuilder, fn RowHandler) error {
	query, args, err := qry.ToSql()
	if err != nil {
		return err
	}
	rows, err := db.Queryx(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := fn(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

