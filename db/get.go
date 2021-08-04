package dbutil

import (
	"fmt"

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

func Exists(db sqlx.Queryer, qry squirrel.SelectBuilder) (bool, error) {
	var exists bool
	if err := Get(db, &exists, qry.Prefix("SELECT EXISTS(").Suffix(")")); err != nil {
		return false, err
	}
	return exists, nil
}

func MkJoinStr(t1, c1, t2, c2 string) string {
	return fmt.Sprintf("%s ON %[1]s.%[2]s = %[3]s.%[4]s", t1, c1, t2, c2)
}