package dbUtil

import (
	"reflect"
	"strings"

	struct_tag "github.com/monstercat/golib/struct-tag"
)

// CUSTOM-MAP tags should be avoided and we should be using JSON tags where possible
func decodeFieldName(f reflect.StructField) string {
	if customTag := f.Tag.Get("custom-map"); customTag != "" {
		return customTag
	}

	if tag := f.Tag.Get("json"); tag != "" {
		return tag
	}

	return ""
}

// GetPayloadMapping takes a dataModel being modified and a payload to modify it
// it returns a map that can be used in a SetMap to modify database values
func GetPayloadMapping(dataModel interface{}, payload map[string]interface{}) map[string]interface{} {
	set := make(map[string]interface{})

	struct_tag.IterateStructFields(dataModel, func(f reflect.StructField, v reflect.Value) {
		name := decodeFieldName(f)
		if name == "" || name == "-" {
			return
		}

		dbName := ColumnNameFromDbTag(f)
		if dbName == "" {
			return
		}

		val, ok := payload[name]
		if !ok {
			return
		}

		typ := f.Type.String()
		if strings.Contains(typ, "pgnull") || strings.Contains(typ, "sql.Null") {
			set[dbName] = val
			return
		}

		if val == nil || f.Type != reflect.TypeOf(val) {
			return
		}

		set[dbName] = val
	})

	return set
}
