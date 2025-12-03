package postgres

import "github.com/Masterminds/squirrel"

// Eq is a type-alias for squirrel.Eq. It adds onto squirrel.Eq's SQL string
// by encapsulating it with brackets. This will cause it to work well with
// squirrel.Or.
type Eq squirrel.Eq

func (eq Eq) ToSql() (string, []interface{}, error) {
	sql, args, err := squirrel.Eq(eq).ToSql()
	if err != nil {
		return "", nil, err
	}
	return "(" + sql + ")", args, nil
}
