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
	for i := 0; i < refVal.NumField(); i++ {
		it(refVal.Type().Field(i), refVal.Field(i))
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
	case time.Time:
		isZero = dt.IsZero()
		data = dt
		checked = true
	}
	return
}
