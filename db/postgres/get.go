package pgUtils

import (
	"github.com/jmoiron/sqlx"

	. "github.com/monstercat/golib/db"
)

func GetForStruct(db sqlx.Queryer, val interface{}, table string, where interface{}) error {
	cols := GetColumnsList(val, "")
	return Get(db, val, Psql.Select(cols...).From(table).Where(where))
}

func SelectForStruct(db sqlx.Queryer, slice interface{}, table string, where interface{}) error {
	cols := GetColumnsList(slice, "")
	return Select(db, slice, Psql.Select(cols...).From(table).Where(where))
}

func Exists(db sqlx.Queryer, table string, where interface{}) (bool, error) {
	var rows int
	sql, args, err := Psql.Select("COUNT(*)").
		From(table).
		Where(where).ToSql()
	if err != nil {
		return false, err
	}

	if err := sqlx.Get(db, &rows, sql, args...); err != nil {
		return false, err
	}
	return rows > 0, nil
}

