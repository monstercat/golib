package struct_tag

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"time"

	"github.com/monstercat/pgnull"
)

type StructFieldIterator func(reflect.StructField, reflect.Value)

func IterateStructFields(e interface{}, it StructFieldIterator) {
	refVal := reflect.ValueOf(e).Elem()
	refType := reflect.TypeOf(e).Elem()

	var numField int
	// If the element is a slice, we get the underlying element.
	// If the underlying element is a pointer, we remove the pointer (get the underlying element).
	// This is required to get the proper reflect Value to iterate on.
	// Otherwise, NumField and Field do not exist. 
	if refType.Kind() == reflect.Slice {
		refType = refType.Elem()
		if refType.Kind() == reflect.Ptr {
			refType = refType.Elem()
		}
		numField = refType.NumField()

		// Also need to change refVal.
		refVal = reflect.Indirect(reflect.New(refType))
	} else if refType.Kind() == reflect.Interface {
		refVal = refVal.Elem()
		refType = refVal.Type()
		numField = refVal.NumField()
	} else {
		numField = refVal.NumField()
	}

	for i := 0; i < numField; i++ {
		it(refType.Field(i), refVal.Field(i))
	}
}

func GetData(v reflect.Value) (data interface{}, isZero bool, checked bool) {
	data = v.Interface()
	isZero = v.IsZero()

	switch dt := data.(type) {
	case map[string]interface{}:
		isZero = len(dt) == 0
		var err error
		data, err = json.Marshal(dt)
		checked = err == nil
	case sql.NullString:
		isZero = !dt.Valid || dt.String == ""
		checked = true
	case pgnull.NullString:
		isZero = !dt.Valid || dt.String == ""
		checked = true
	case pgnull.NullTime:
		isZero = dt.Time.IsZero()
		checked = true
	case pgnull.NullInt:
		isZero = !dt.Valid || dt.Int64 == 0
		checked = true
	case pgnull.NullFloat:
		isZero = !dt.Valid || dt.Float64 == 0
		checked = true
	case sql.NullInt32:
		isZero = !dt.Valid || dt.Int32 == 0
		checked = true
	case sql.NullInt64:
		isZero = !dt.Valid || dt.Int64 == 0
		checked = true
	case sql.NullBool:
		isZero = !dt.Valid || !dt.Bool
		checked = true
	case time.Time:
		isZero = dt.IsZero()
		data = dt
		checked = true
	}
	return
}
