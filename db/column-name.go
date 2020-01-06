package dbUtil

import (
	"github.com/monstercat/golib/string"
	"reflect"
	"strings"
)

func ColumnNameFromDbTag(f reflect.StructField) string {
	dbTag := f.Tag.Get("db")
	if dbTag == "-" {
		return ""
	}
	if dbTag == "" {
		return stringutil.CamelToSnakeCase(f.Name)
	}

	str :=  strings.Split(dbTag, ",")
	if len(str) == 0 {
		return ""
	}
	return str[0]
}
