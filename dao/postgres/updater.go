package postgres

import (
	"errors"

	"github.com/Masterminds/squirrel"

	"github.com/monstercat/golib/dao/daohelpers"
)

var (
	ErrMissingProvider = errors.New("missing db provider")
)

// Updater provides some default insert/update/delete functionality. The parameter T is the type for the ID which is
// returned through the insert.
type Updater[T any] struct {
	// The base query builder object. Needs to be passed in order to properly generate the query.
	QueryBuilder *UpdateBuilder

	// What provides the DB connection for calling the functions.
	Provider DBProvider

	// PreprocessUpdate preprocesses the update query
	PreprocessUpdate Preprocessor[squirrel.UpdateBuilder]

	// PreprocessInsert preprocesses the insert query
	PreprocessInsert Preprocessor[squirrel.InsertBuilder]

	// PreprocessDelete preprocesses the delete query
	PreprocessDelete Preprocessor[squirrel.DeleteBuilder]

	// Set of data to update with.
	Data map[string]interface{}

	// Any error through SET logic.
	err error
}

// SetError can be used by encapsulating structs to capture errors during set.
// For example, if JSON marshalling is required, the error return value of
// json.Marshal can be passed to this function.
func (u *Updater[T]) SetError(err error) {
	u.err = err
}

func NewUpdater[T any](builder *StatementBuilder, baseTable string) *Updater[T] {
	updateBuilder := NewUpdateBuilder(builder).SetBaseTable(baseTable)
	return &Updater[T]{
		QueryBuilder: updateBuilder,
		Data:         make(map[string]interface{}),
	}
}

func (u *Updater[T]) SetProvider(db DBProvider) *Updater[T] {
	u.Provider = db
	return u
}

func (u *Updater[T]) SetIdColumn(col string) *Updater[T] {
	u.QueryBuilder.IdColumnName = col
	return u
}

func (u *Updater[T]) Set(name string, value interface{}) {
	u.Data[name] = value
}

func (u *Updater[T]) Update() error {
	if u.Provider == nil {
		return ErrMissingProvider
	}
	if !u.QueryBuilder.HasConditions() {
		return daohelpers.ErrNoConditions
	}
	// Return an error if an error is present.
	if u.err != nil {
		return u.err
	}
	qry := u.QueryBuilder.UpdateBuilder(u.Data)
	if u.PreprocessUpdate != nil {
		qry = u.PreprocessUpdate(u.QueryBuilder.table, qry)
	}
	res, err := qry.
		PlaceholderFormat(squirrel.Dollar).
		RunWith(u.Provider.GetDb()).
		Exec()
	if err != nil {
		return err
	}

	// We need to check results! If nothing was updated, an error should be
	// returned.
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return daohelpers.ErrNoUpdatePerformed
	}
	return nil
}

func (u *Updater[T]) Insert() (T, error) {
	var id T
	err := u.QueryBuilder.InsertBuilder(u.Data).
		PlaceholderFormat(squirrel.Dollar).
		Suffix("RETURNING " + u.QueryBuilder.IdColumnName).
		RunWith(u.Provider.GetDb()).
		Scan(&id)
	return id, err
}

func (u *Updater[T]) InsertNoId() error {
	_, err := u.QueryBuilder.InsertBuilder(u.Data).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(u.Provider.GetDb()).
		Exec()
	return err
}

func (u *Updater[T]) Delete() error {
	if !u.QueryBuilder.HasConditions() {
		return daohelpers.ErrNoConditions
	}

	qry := u.QueryBuilder.DeleteBuilder()
	if u.PreprocessDelete != nil {
		qry = u.PreprocessDelete(u.QueryBuilder.table, qry)
	}

	res, err := qry.
		PlaceholderFormat(squirrel.Dollar).
		RunWith(u.Provider.GetDb()).
		Exec()
	if err != nil {
		return err
	}

	// We need to check results! If nothing was updated, an error should be
	// returned.
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return daohelpers.ErrNoDeletePerformed
	}
	return nil
}
