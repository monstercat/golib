package postgres

import (
	"strings"

	"github.com/Masterminds/squirrel"
)

// With implements the squirrel.Sqlizer interface for WITH as a prefix.
type With map[string]squirrel.Sqlizer

func (w With) ToSql() (string, []interface{}, error) {
	if len(w) == 0 {
		return "", []interface{}{}, nil
	}
	var sqlParts []string
	var args []interface{}
	for k, s := range w {
		p, a, err := s.ToSql()
		if err != nil {
			return "", nil, err
		}
		sqlParts = append(sqlParts, k+" AS ("+p+")")
		args = append(args, a...)
	}
	return "WITH " + strings.Join(sqlParts, ", "), args, nil
}
