package pgUtils

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	. "github.com/monstercat/golib/db"
)

var Psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func Delete(db sqlx.Ext, qry squirrel.DeleteBuilder) error {
	_, err := qry.RunWith(db).Exec()
	return err
}

func DeleteWhere(db sqlx.Ext, table string, where interface{}) error {
	return Delete(db, Psql.Delete(table).Where(where))
}

func Update(db sqlx.Ext, qry squirrel.UpdateBuilder) error {
	_, err := qry.RunWith(db).Exec()
	return err
}

func UpdateSetMap(db sqlx.Ext, table string, payload, where interface{}) error {
	return Update(db, Psql.Update(table).SetMap(SetMap(payload, false)).Where(where))
}

func InsertReturningId(db sqlx.Ext, qry squirrel.InsertBuilder, id interface{}) error {
	return qry.Suffix("RETURNING id").RunWith(db).Scan(id)
}

func InsertSetMapReturningId(db sqlx.Ext, table string, payload interface{}, id interface{}) error {
	return InsertReturningId(db, Psql.Insert(table).SetMap(SetMap(payload, true)), id)
}

func InsertSetMapNoId(db sqlx.Ext, table string, payload interface{}) error {
	_, err := Psql.Insert(table).SetMap(SetMap(payload, true)).RunWith(db).Exec()
	return err
}