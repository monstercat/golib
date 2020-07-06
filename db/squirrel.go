package dbUtil

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
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
