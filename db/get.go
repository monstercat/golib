package dbUtil

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func Get(db sqlx.Queryer, dest interface{}, qry squirrel.SelectBuilder) error {
	sql, args, err := qry.ToSql()
	if err != nil {
		return err
	}
	return sqlx.Get(db, dest, sql, args...)
}

func Select(db sqlx.Queryer, dest interface{}, qry squirrel.SelectBuilder) error {
	sql, args, err := qry.ToSql()
	if err != nil {
		return err
	}
	return sqlx.Select(db, dest, sql, args...)
}
