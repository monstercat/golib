package dbUtil

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func Get(db sqlx.Queryer, dest interface{}, qry squirrel.SelectBuilder) error {
	sql, args, err := qry.ToSql()
	if err != nil {
		return err
	}
	return IgnoreNoRows(sqlx.Get(db, dest, sql, args...))
}

func Select(db sqlx.Queryer, dest interface{}, qry squirrel.SelectBuilder) error {
	sql, args, err := qry.ToSql()
	if err != nil {
		return err
	}
	return IgnoreNoRows(sqlx.Select(db, dest, sql, args...))
}

func IgnoreNoRows(err error) error {
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}

