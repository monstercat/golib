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
