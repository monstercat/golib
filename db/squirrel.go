package dbUtil

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/squirrel"
)

var (
	ErrNoSelects = errors.New("no selects")
)

type ILikeAnd map[string]interface{}

func (il ILikeAnd) ToSql() (string, []interface{}, error) {
	return getIlikeSql(il, " AND ")
}

type ILikeOr map[string]interface{}

func (il ILikeOr) ToSql() (string, []interface{}, error) {
	return getIlikeSql(il, " OR ")
}

func getIlikeSql(vals map[string]interface{}, join string) (sql string, args []interface{}, err error) {
	var exprs []string
	for key, val := range vals {
		if v, ok := val.(driver.Valuer); ok {
			if val, err = v.Value(); err != nil {
				return
			}
		}
		if val == nil {
			continue
		}
		if isListType(val) {
			continue
		}
		exprs = append(exprs, fmt.Sprintf("%s ILIKE ?", key))
		args = append(args, val)
	}
	sql = "(" + strings.Join(exprs, join) + ")"
	return
}

func isListType(val interface{}) bool {
	if driver.IsValue(val) {
		return false
	}
	valVal := reflect.ValueOf(val)
	return valVal.Kind() == reflect.Array || valVal.Kind() == reflect.Slice
}

type InvalidUnionError struct {
	BaseError error
	Index     int
}

func (e InvalidUnionError) Error() string {
	return fmt.Sprintf("[%d]: %s", e.Index, e.BaseError)
}

// Allows for unions queries. Created another select query using suffixes. Where and Join queries should not be made
// after calling this query, as those queries will only affect the *first* query on the list.
func Union(selects ...squirrel.SelectBuilder) (squirrel.SelectBuilder, error) {
	fail := func(err error) (squirrel.SelectBuilder, error) {
		return squirrel.SelectBuilder{}, err
	}

	if len(selects) == 0 {
		return fail(ErrNoSelects)
	}
	if len(selects) == 1 {
		return selects[0], nil
	}

	// TODO: check columns to match

	first := selects[0]
	selects = selects[1:]

	for i, s := range selects {
		sql, args, err := s.ToSql()
		if err != nil {
			return fail(InvalidUnionError{
				BaseError: err,
				Index:     i,
			})
		}
		first = first.Prefix("("+sql+") UNION", args...)
	}
	return first, nil
}
