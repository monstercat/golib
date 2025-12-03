package postgres

import (
	"github.com/Masterminds/squirrel"
	"github.com/cyc-ttn/go-collections"

	"github.com/monstercat/golib/daohelpers"
)

// ColumnTranslator translates a column from C to a proper string.
type ColumnTranslator[C ~string] func(col C) string

// ScannerFuncFactory generates scanner functions depending on the columns
// required.
type ScannerFuncFactory[T any, C ~string] func(cols ...C) ScannerFunc[T]

// TODO: allow preprocessing of the query based on the columns. We can do this
//   cloning the selector, the QueryBuilder and the StatementBuilder to add
//   joins as necessary

// CustomColumnSelector allows selection of columns during Get, Select and
// Iterate methods. It may reduce the need to join tables, thereby allowing
// performance gains
type CustomColumnSelector[T any, C ~string] struct {
	*Selector[T]

	// DefaultCols is the default list columns to be retrieved by Select or Get.
	// This is in case no columns are provided.
	DefaultCols []C

	// ColumnTranslator is used to translate the columns defined in C into a
	// proper string. If the returned string is an empty string, it will be
	// ignored.
	ColumnTranslator ColumnTranslator[C]

	// ScannerFactory generates ScannerFunc methods used for scanning data. This
	// is optional (in case intermediate steps are required). By default, it will
	// attempt to use the T struct directly (bigquery provides reflection by
	// default).
	//
	// However, if struct fields are to be used to dictate how bigquery's
	// reflection works, either a private type should be provided, and the field
	// values copied to T, or a custom bigquery.ValueLoader which loads values
	// directly into T
	ScannerFactory ScannerFuncFactory[T, C]

	// Preprocess allows modification of the query based on the column before
	// selection has occurred. This can be used to add joins or prefixes.
	Preprocess func(tbl string, qry squirrel.SelectBuilder, cols ...C) squirrel.SelectBuilder
}

// preprocessColumns preprocesses the columns and converts them to a list of
// table columns with a corresponding ScannerFunc. An error will be returned if
// no columns are provided.
func (s *CustomColumnSelector[T, C]) preprocessColumns(cols ...C) (ScannerFunc[T], []string, error) {
	if len(cols) == 0 {
		return nil, nil, ErrSelectorMissingColumns
	}

	// Generate the route that scans based on the columns.
	scanner := s.ScannerFactory(cols...)

	// Convert the columns into a list of strings.
	colNames := collections.Map[string](cols, func(agg []string, col C) (string, bool) {
		name := s.ColumnTranslator(col)
		return name, name != ""
	})

	return scanner, colNames, nil
}

// Get returns a single item
func (s *CustomColumnSelector[T, C]) Get(cols ...C) (T, error) {
	if len(cols) == 0 {
		cols = s.DefaultCols
	}
	scanner, colNames, err := s.preprocessColumns(cols...)
	if err != nil {
		var t T
		return t, ErrSelectorMissingColumns
	}
	defer s.preprocessGet(cols...)()
	return Get[T](s.Selector, scanner, colNames...)
}

func (s *CustomColumnSelector[T, C]) preprocessSelect(cols ...C) func() {
	if s.Preprocess == nil {
		return func() {}
	}
	oldPreprocessor := s.PreprocessSelect
	s.PreprocessSelect = func(tbl string, qry squirrel.SelectBuilder) squirrel.SelectBuilder {
		qry = s.Preprocess(tbl, qry, cols...)
		if oldPreprocessor != nil {
			return oldPreprocessor(tbl, qry)
		}
		return qry
	}
	return func() {
		s.PreprocessSelect = oldPreprocessor
	}
}

func (s *CustomColumnSelector[T, C]) preprocessGet(cols ...C) func() {
	if s.Preprocess == nil {
		return func() {}
	}
	oldPreprocessor := s.PreprocessGet
	s.PreprocessGet = func(tbl string, qry squirrel.SelectBuilder) squirrel.SelectBuilder {
		qry = s.Preprocess(tbl, qry, cols...)
		if oldPreprocessor != nil {
			return oldPreprocessor(tbl, qry)
		}
		return qry
	}
	return func() {
		s.PreprocessGet = oldPreprocessor
	}
}

// Select returns a list of items
func (s *CustomColumnSelector[T, C]) Select(cols ...C) ([]T, error) {
	if len(cols) == 0 {
		cols = s.DefaultCols
	}
	scanner, colNames, err := s.preprocessColumns(cols...)
	if err != nil {
		return nil, ErrSelectorMissingColumns
	}

	// Append the preprocessor.
	defer s.preprocessSelect(cols...)()
	return Select[T](s.Selector, scanner, colNames...)
}

// Iterate returns an object that iterates through a list of results one
// at a time.
func (s *CustomColumnSelector[T, C]) Iterate(cols ...C) (daohelpers.ObjectIterator[T], error) {
	if len(cols) == 0 {
		cols = s.DefaultCols
	}
	scanner, colNames, err := s.preprocessColumns(cols...)
	if err != nil {
		return nil, ErrSelectorMissingColumns
	}

	// Append the preprocessor.
	defer s.preprocessSelect(cols...)()
	return Iterate[T](s.Selector, scanner, colNames...)
}
