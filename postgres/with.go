package postgres

import (
	"strings"

	"github.com/cyc-ttn/go-collections"

	"github.com/Masterminds/squirrel"
)

// With implements the squirrel.Sqlizer interface for WITH as a prefix. Prefixes
// may need to be added in order.
type With struct {
	// If true, will add "RECURSIVE" after "WITH"
	IsRecursive bool
	Keys        []string
	SQL         []squirrel.Sqlizer
}

// Add will add an item to With. *With should not be nil.
func (w *With) Add(name string, sql squirrel.Sqlizer) {
	// Check existence. If exists, ignore.
	if collections.Contains(name, w.Keys) {
		return
	}
	w.Keys = append(w.Keys, name)
	w.SQL = append(w.SQL, sql)
}

// ToSql converts With to SQL. *With can be nil. In that case, an empty SQL
// string / arguments is returned.
func (w *With) ToSql() (string, []interface{}, error) {
	if w == nil || len(w.Keys) == 0 {
		return "", []interface{}{}, nil
	}
	var sqlParts []string
	var args []interface{}
	for idx, k := range w.Keys {
		s := w.SQL[idx]
		p, a, err := s.ToSql()
		if err != nil {
			return "", nil, err
		}
		sqlParts = append(sqlParts, k+" AS ("+p+")")
		args = append(args, a...)
	}
	prefix := "WITH "
	if w.IsRecursive {
		prefix = "WITH RECURSIVE "
	}
	return prefix + strings.Join(sqlParts, ", "), args, nil
}

// Clone will clone a *With. It will handle the case where *With is nil.
func (w *With) Clone() *With {
	if w == nil {
		return &With{}
	}
	return &With{
		Keys: append([]string{}, w.Keys...),
		SQL:  append([]squirrel.Sqlizer{}, w.SQL...),
	}
}
