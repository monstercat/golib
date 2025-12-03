package postgres

import (
	"github.com/Masterminds/squirrel"
	"github.com/cyc-ttn/go-collections"

	"github.com/monstercat/golib/daohelpers"
)

// BatchLoaderContext provides information to custom processors, such as
// AddSuffixFn at the time of Flush.
type BatchLoaderContext struct {
	Columns []string // columns being inserted at the time of flush.
}

// BatchEntryFactory is a factory for generating BatchLoaderEntry objects of a
// specific type.
type BatchEntryFactory[T daohelpers.BatchLoaderEntry] func(bl *BatchLoader[T]) T

// BatchLoader loads (inserts) multiple points of data in a batch. It implements
// daohelpers.BatchLoader
type BatchLoader[T daohelpers.BatchLoaderEntry] struct {
	// SkipConflicts will add "ON CONFLICT DO NOTHING" to the insert query.
	SkipConflicts bool

	// AddSuffixFn is a function that can be used to add a suffix to the insert
	AddSuffixFn func(context *BatchLoaderContext) string

	// What provides the DB connection for calling the functions.
	Provider DBProvider

	// Factory creates new entries.
	Factory BatchEntryFactory[T]

	// Table to insert data into
	Table string

	// List of columns to insert. It assumes that all added data has the same
	// list of columns.
	columns map[string]bool

	// Values added to the batch loader.
	values []map[string]any

	// Default values to use in case they are not provided.
	Defaults map[string]any

	// Maximum number of values added at one time. If 0, it will be treated as
	// infinite.
	MaxValues uint64
}

func NewBatchLoader[T daohelpers.BatchLoaderEntry](
	Provider DBProvider,
	factory BatchEntryFactory[T],
	table string,
) *BatchLoader[T] {
	return &BatchLoader[T]{
		Provider: Provider,
		Factory:  factory,
		Table:    table,
		columns:  make(map[string]bool),
		Defaults: make(map[string]any),
	}
}

func (l *BatchLoader[T]) SetMaxValues(maxValues uint64) *BatchLoader[T] {
	l.MaxValues = maxValues
	return l
}

// SetDefaults sets default values for columns. If not provided, NIL will be
// used instead.
func (l *BatchLoader[T]) SetDefaults(def map[string]any) *BatchLoader[T] {
	for k := range def {
		l.columns[k] = true
	}
	l.Defaults = def
	return l
}

// Add adds the provided values to the set of values. This can be called
// for example, by the BatchLoaderEntry.Commit method. It expects a map of
// values keyed by column name.
func (l *BatchLoader[T]) Add(val map[string]any) error {
	for k := range val {
		l.columns[k] = true
	}
	l.values = append(l.values, val)

	// If we have hit the maximum number of values, we should send the data.
	if l.MaxValues > 0 && uint64(len(l.values)) > l.MaxValues {
		vals := l.values[l.MaxValues:]
		l.values = l.values[0:l.MaxValues]
		if err := l.Flush(); err != nil {
			return err
		}
		l.values = vals
	}

	return nil
}

// New creates a new BatchLoaderEntry.
func (l *BatchLoader[T]) New() T {
	return l.Factory(l)
}

// Returns a list of columns. This list will be used to place all values in the
// correct order.
func (l *BatchLoader[T]) colList() []string {
	cols := make([]string, 0, len(l.columns))
	for k := range l.columns {
		cols = append(cols, k)
	}
	return cols
}

// Len should return the number of entries not yet flushed.
func (l *BatchLoader[T]) Len() int {
	return len(l.values)
}

// Flush should commit any fully loaded data to the backend.
func (l *BatchLoader[T]) Flush() error {
	cols := l.colList()

	// TODO convert to using copy.
	insQry := squirrel.Insert(l.Table).
		Columns(cols...)
	for _, value := range l.values {
		vals := collections.Map[any](cols, func(agg []any, s string) (any, bool) {
			v, exists := value[s]
			if exists {
				return v, true
			}
			v, exists = l.Defaults[s]
			if exists {
				return v, true
			}
			return nil, true
		})

		insQry = insQry.Values(vals...)
	}

	// Reset so that Len works properly.
	l.values = nil

	if l.SkipConflicts {
		insQry = insQry.Suffix("ON CONFLICT DO NOTHING")
	}
	if l.AddSuffixFn != nil {
		insQry = insQry.Suffix(l.AddSuffixFn(&BatchLoaderContext{
			Columns: cols,
		}))
	}

	_, err := insQry.
		RunWith(l.Provider.GetDb()).
		PlaceholderFormat(squirrel.Dollar).
		Exec()
	return err
}
