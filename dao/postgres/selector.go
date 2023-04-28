package postgres

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"google.golang.org/api/iterator"
)

var (
	ErrSelectorMissingColumns = errors.New("columns are missing")
)

// ScannerFunc is a definition for a function that performs scanning.
type ScannerFunc[T any] func(scanner squirrel.RowScanner) (T, error)

func (f ScannerFunc[T]) Scan(scanner squirrel.RowScanner) (T, error) {
	return f(scanner)
}

// Scanner is an object that contains a Scan method. The method expects a row of data From the source database to be
// used to create a single object of type T.
type Scanner[T any] interface {
	Scan(scanner squirrel.RowScanner) (T, error)
}

// DBProvider is an interface that provides the database connection to the Selector.
type DBProvider interface {
	GetDb() sqlx.Ext
}

// Preprocessor is a generic type for preprocessing queries.
type Preprocessor[T squirrel.Sqlizer] func(tbl string, sql T) T

// Selector provides three methods that can be used as defaults.
// 1. Total - returns the total # of objects that satisfies the query.
// 2. Select - returns a list of results
// 3. Get - returns a single result.
type Selector[T any] struct {
	// The base query builder object. Needs to be passed in order to properly generate the query.
	QueryBuilder *SelectBuilder

	// The scanner that outputs an object containing the values of the row. This is used for both Get and Select.
	// The columns scanned by the Scanner *must* match the columns provided in GetCols.
	Scanner Scanner[T]

	// What provides the DB connection for calling the functions.
	Provider DBProvider

	// Columns related to Get and Select
	GetCols []string

	// Column related to the Total
	TotalColumnName string

	// PreprocessSelect allows preprocessing of the select query *before*
	// Querying and scanning. It, for example, allows for adding of extra
	// conditions or columns.
	//
	// The same preprocessor is used for Iterate, Exists, and Total
	PreprocessSelect func(from string, sql squirrel.SelectBuilder) squirrel.SelectBuilder

	// PreprocessGet allows preprocessing of the get query *before* Querying
	// and scanning. It, for example, allows for adding of extra
	// conditions or columns.
	PreprocessGet func(builder squirrel.SelectBuilder) squirrel.SelectBuilder

	// ProcessSelectResult allows further processing as a result of the result
	// of the select.
	ProcessSelectResult func(val interface{}, err error)

	// ProcessGetResult allows further processing as a result of the result
	// of the get.
	ProcessGetResult func(val interface{}, err error) error
}

// Get returns a single object that satisfies the query, up to a certain Limit. It handles sorting but paging is
// unncessary as the first object will always be the one that is returned. If no columns are provided,
// ErrSelectorMissingColumns is returned.
func (s *Selector[T]) Get() (T, error) {
	if len(s.GetCols) == 0 {
		var t T
		return t, ErrSelectorMissingColumns
	}
	return Get[T](s, s.Scanner, s.processGetCols()...)
}

func (s *Selector[T]) processGetCols() []string {
	c := make([]string, 0, len(s.GetCols))
	for _, col := range s.GetCols {
		c = append(c, strings.Replace(col, TablePlaceholder, s.QueryBuilder.From, -1))
	}
	return c
}

// Get returns a single object with custom columns provided.
func Get[T any](s *Selector[T], scanner Scanner[T], cols ...string) (T, error) {
	qry := s.QueryBuilder.Builder(cols...).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.Provider.GetDb())
	s.QueryBuilder.ApplySort(&qry)

	// Add preprocessor
	if s.PreprocessGet != nil {
		qry = s.PreprocessGet(qry)
	}

	fn := func(val T, err error) (T, error) {
		if s.ProcessGetResult != nil {
			err = s.ProcessGetResult(val, err)
		}
		return val, err
	}
	return fn(scanner.Scan(qry))
}

// Select returns a list of objects that satisfies the query, up to a certain Limit. It also handles sorting
// and Offset. Columns are required for the selector to function. If no columns are provided, ErrSelectorMissingColumns
// is returned.
func (s *Selector[T]) Select() ([]T, error) {
	if len(s.GetCols) == 0 {
		return nil, ErrSelectorMissingColumns
	}
	return Select[T](s, s.Scanner, s.processGetCols()...)
}

// Select returns a list of objects with columns provided.
// It uses the conditions / paging / sort in the provided
// Selector to restrict the output.
func Select[R any](s *Selector[R], scanner Scanner[R], cols ...string) ([]R, error) {
	fn := func(val []R, err error) ([]R, error) {
		if s.ProcessSelectResult != nil {
			s.ProcessSelectResult(val, err)
		}
		return val, err
	}

	rows, err := Iterate(s, scanner, cols...)
	if err != nil {
		return fn(nil, err)
	}

	var xs []R
	for {
		obj, err := rows.Next()
		if err == iterator.Done {
			return fn(xs, nil)
		}
		if err != nil {
			return fn(nil, err)
		}

		xs = append(xs, obj)
	}
}

// Iterate returns an iterator to retrieve objects with the columns provided.
// It uses the conditions / paging / sort in the provided
// Selector to restrict the output.
func (s *Selector[T]) Iterate() (*SelectIterator[T], error) {
	if len(s.GetCols) == 0 {
		return nil, ErrSelectorMissingColumns
	}
	return Iterate[T](s, s.Scanner, s.processGetCols()...)
}

// Iterate returns an iterator for retrieving multiple rows sequentially From the database.
func Iterate[R any](s *Selector[R], scanner Scanner[R], cols ...string) (*SelectIterator[R], error) {
	qry := s.QueryBuilder.Builder(cols...).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.Provider.GetDb())
	s.QueryBuilder.ApplyPaging(&qry)
	s.QueryBuilder.ApplySort(&qry)

	// Add preprocessor
	if s.PreprocessSelect != nil {
		qry = s.PreprocessSelect(s.QueryBuilder.From, qry)
	}

	rows, err := qry.Query()
	if err != nil {
		return nil, err
	}

	return &SelectIterator[R]{
		Rows: rows,
		Fn:   scanner,
	}, nil
}

// Total returns the total # of objects that satisfies the query. It will extract the query form the StatementBuilder
// and apply conditions to it. It attempts to extract the column From TotalColumnName, but will default to COUNT(*)
// if it is not provided.
func (s *Selector[T]) Total() (uint64, error) {
	col := s.TotalColumnName
	if col == "" {
		col = "COUNT(*)"
	}

	qry := s.QueryBuilder.Builder(col).
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.Provider.GetDb())

	// Add preprocessor
	if s.PreprocessSelect != nil {
		qry = s.PreprocessSelect(s.QueryBuilder.From, qry)
	}
	return SingleColumnScanner[uint64](qry)
}

// Exists returns true if any result filtered by the query is present.
func (s *Selector[T]) Exists() (bool, error) {
	qry := s.QueryBuilder.Builder("*").
		PlaceholderFormat(squirrel.Dollar).
		RunWith(s.Provider.GetDb()).
		Prefix("SELECT EXISTS(").
		Suffix(")")

	// Add preprocessor
	if s.PreprocessSelect != nil {
		qry = s.PreprocessSelect(s.QueryBuilder.From, qry)
	}
	return SingleColumnScanner[bool](qry)
}

func SingleColumnScanner[T any](row squirrel.RowScanner) (T, error) {
	var t T
	err := row.Scan(&t)
	return t, err
}

// SelectIterator is a way to retrieve an iterator for a select to use with
// scanning.
type SelectIterator[R any] struct {
	Rows *sql.Rows // Rows to iterate over.
	Fn   Scanner[R]
}

// Next returns the next item in the list.
func (i *SelectIterator[R]) Next() (R, error) {
	if !i.Rows.Next() {
		var r R
		return r, iterator.Done
	}
	return i.Fn.Scan(i.Rows)
}
