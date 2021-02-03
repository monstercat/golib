package dbUtil

import (
	"reflect"
	"strings"

	structTag "github.com/monstercat/golib/struct-tag"
)

// MapPayloadToColumns takes a payload and maps it to the correct table values
// this can be used in a SetMap to set values
func MapPayloadToColumns(payload interface{}) map[string]interface{} {
	pyld := make(map[string]interface{})

	structTag.IterateStructFields(payload, func(field reflect.StructField, value reflect.Value) {
		if value.IsNil() {
			return
		}

		tagStr := field.Tag.Get("column-mapping")

		if tagStr == "" {
			tagStr = field.Name
		}

		tags := strings.Split(tagStr, ",")

		for _, v := range tags {
			pyld[v] = value.Elem()
		}
	})

	return pyld
}

// GetModified takes a payload struct with column-mappings and checks for values
// that can be updated in the original object
// func GetModified(payload interface{}, original interface{}) map[string]interface{} {
// 	pyld := MapPayloadToColumns(payload)
// 	modified := make(map[string]interface{})

// 	structTag.IterateStructFields(original, func(field reflect.StructField, value reflect.Value) {
// 		name := ColumnNameFromDbTag(field)
// 		fmt.Printf("\nNAME" + name)
// 		val, ok := pyld[name]
// 		fmt.Printf("\nVAL%v OK!!! %t", val, ok)
// 		if !ok {
// 			return
// 		}

// 		typ := field.Type.String()
// 		fmt.Printf("\nTYPE%s", value)
// 		if strings.Contains(typ, "pgnull") || strings.Contains(typ, "sql.Null") {
// 			modified[name] = val
// 			return
// 		}

// 		if field.Type != reflect.TypeOf(val) {
// 			return
// 		}
// 		modified[name] = val
// 	})
// 	fmt.Printf("\nMODIFIED:: %+v", modified)

// 	return modified
// }
