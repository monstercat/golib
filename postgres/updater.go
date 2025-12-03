package postgres

import (
	"errors"

	"github.com/Masterminds/squirrel"

	"github.com/monstercat/golib/daohelpers"
)

var (
	ErrMissingConditions = errors.New("missing conditions")
	ErrMissingProvider   = errors.New("missing db provider")
)

// Updater provides some default insert/update/delete functionality. The parameter T is the type for the ID which is
// returned through the insert.
type Updater[T any] struct {
	// The base query builder object. Needs to be passed in order to properly generate the query.
	QueryBuilder *UpdateBuilder

	// What provides the DB connection for calling the functions.
	Provider DBProvider

	// Set of data to update with.
	Data map[string]interface{}

	// Override provides override values for specified columns. If a replacement,
	// value exists in the map, it will be assigned to Data map when Set is
	// called on the column.
	Override map[string]interface{}

	// Constraint for upsert. Required for upsert to work. For example:
	// - ON CONSTRAINT [constraint name]
	// - ([column name])
	UpsertConstraint string

	// PreprocessUpdate preprocesses the update query
	PreprocessUpdate Preprocessor[squirrel.UpdateBuilder]

	// PreprocessInsert preprocesses the insert query
	PreprocessInsert Preprocessor[squirrel.InsertBuilder]

	// PreprocessDelete preprocesses the delete query
	PreprocessDelete Preprocessor[squirrel.DeleteBuilder]

	// Any error through SET logic.
	err error
}

func NewUpdater[T any](builder *StatementBuilder, baseTable string) *Updater[T] {
	updateBuilder := NewUpdateBuilder(builder).SetBaseTable(baseTable)
	return &Updater[T]{
		QueryBuilder: updateBuilder,
		Data:         make(map[string]interface{}),
		Override:     make(map[string]interface{}),
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

func (u *Updater[T]) SetUpsertCondition(cond string) *Updater[T] {
	u.UpsertConstraint = cond
	return u
}

func (u *Updater[T]) SetUpsertConstraints(constraints string) *Updater[T] {
	u.UpsertConstraint = "ON CONSTRAINT " + constraints
	return u
}

// SetError can be used by encapsulating structs to capture errors during set.
// For example, if JSON marshalling is required, the error return value of
// json.Marshal can be passed to this function.
func (u *Updater[T]) SetError(err error) {
	u.err = err
}

func (u *Updater[T]) GetError() error {
	return u.err
}

// SetOverrides sets the override values for specified columns which will be
// used in Set updating columns instead of the given values.
func (u *Updater[T]) SetOverrides(overrides map[string]any) {
	u.Override = overrides
	if len(u.Data) > 0 {
		for k, v := range overrides {
			u.Set(k, v)
		}
	}
}

// Set adds a value to the data set.
// Supports partial update by checking if the value is ommitted.
// Supports nilable.Value type if Nullable option is specified in opts.
func (u *Updater[T]) Set(name string, value interface{}, opts ...daohelpers.UpdateRule) {
	var interuppt bool
	if v, ok := u.Override[name]; ok {
		value = v
	} else {
		for _, option := range opts {
			value, interuppt = option(name, value)
			if interuppt {
				u.Data[name] = value
				return
			}
		}
	}

	if IsValueOmitted(value) {
		u.Remove(name)
		return
	}
	u.Data[name] = value
}

// Remove removes a value from the data set
func (u *Updater[T]) Remove(name string) {
	delete(u.Data, name)
}

func (u *Updater[T]) GetStringArrayFromData(key string) ([]string, bool) {
	i, ok := u.Data[key]
	if !ok {
		return nil, false
	}
	var arr []string
	arr, ok = i.([]string)
	if !ok {
		return nil, false
	}
	return arr, true
}

func (u *Updater[T]) GetStringFromData(key string) (string, bool) {
	i, ok := u.Data[key]
	if !ok {
		return "", false
	}
	s, ok := i.(string)
	if !ok {
		return "", false
	}
	return s, true
}

// Update will run an update query.
func (u *Updater[T]) Update() error {
	if u.Provider == nil {
		return ErrMissingProvider
	}
	if !u.QueryBuilder.HasConditions() {
		return ErrMissingConditions
	}

	// Return an error if an error is present.
	if u.err != nil {
		return u.err
	}

	_, err := u.CreateUpdateQuery().
		PlaceholderFormat(squirrel.Dollar).
		RunWith(u.Provider.GetDb()).
		Exec()
	return err
}

// CreateUpdateQuery creates an update query without changing the placeholders.
// This allows other objects to retrieve the query for use in their own specific
// versions of update.
func (u *Updater[T]) CreateUpdateQuery() squirrel.UpdateBuilder {
	qry := u.QueryBuilder.UpdateBuilder(u.Data)
	if u.PreprocessUpdate != nil {
		qry = u.PreprocessUpdate(u.QueryBuilder.table, qry)
	}
	return qry
}

// CreateInsertQuery creates an insert query without any suffix or changing
// the placeholders. This allows other objects to retrieve the query for use
// in their own specific versions of insert.
func (u *Updater[T]) CreateInsertQuery() squirrel.InsertBuilder {
	qry := u.QueryBuilder.InsertBuilder(u.Data)
	if u.PreprocessInsert != nil {
		qry = u.PreprocessInsert(u.QueryBuilder.table, qry)
	}
	return qry
}

// Insert will run an insert query, returning an ID.
func (u *Updater[T]) Insert() (T, error) {
	// Return an error if an error is present.
	if u.err != nil {
		var t T
		return t, u.err
	}

	var id T
	err := u.CreateInsertQuery().
		PlaceholderFormat(squirrel.Dollar).
		Suffix("RETURNING " + u.QueryBuilder.IdColumnName).
		RunWith(u.Provider.GetDb()).
		Scan(&id)
	return id, err
}

// SilentInsert inserts the data without returning the ID.
func (u *Updater[T]) SilentInsert() error {
	// Return an error if an error is present.
	if u.err != nil {
		return u.err
	}

	_, err := u.CreateInsertQuery().
		PlaceholderFormat(squirrel.Dollar).
		RunWith(u.Provider.GetDb()).
		Exec()
	return err
}

// Delete will run a delete query.
func (u *Updater[T]) Delete() error {
	qry := u.QueryBuilder.DeleteBuilder()
	if u.PreprocessDelete != nil {
		qry = u.PreprocessDelete(u.QueryBuilder.table, qry)
	}

	_, err := qry.
		PlaceholderFormat(squirrel.Dollar).
		RunWith(u.Provider.GetDb()).
		Exec()
	return err
}

// Upsert attempts to insert. On conflict, update will be executed instead.
func (u *Updater[T]) Upsert() error {
	fields := make([]string, 0, len(u.Data))
	for k, _ := range u.Data {
		fields = append(fields, k)
	}
	onConflict := OnConflictUpdateSuffixWithoutConstraint(u.UpsertConstraint, fields...)

	// Create insert query and perform upsert.
	_, err := u.CreateInsertQuery().
		Suffix(onConflict).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(u.Provider.GetDb()).
		Exec()
	return err
}
