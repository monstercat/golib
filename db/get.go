package dbUtil

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func Get(db sqlx.Queryer, val interface{}, qry squirrel.SelectBuilder) error {
	sql, args, err := qry.ToSql()
	if err != nil {
		return err
	}
	return sqlx.Get(db, val, sql, args...)
}

func Select(db sqlx.Queryer, slice interface{}, qry squirrel.SelectBuilder) error {
	sql, args, err := qry.ToSql()
	if err != nil {
		return err
	}
	return sqlx.Select(db, slice, sql, args...)
}
