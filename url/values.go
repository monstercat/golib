package url

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/monstercat/golib/errors"
	structTag "github.com/monstercat/golib/struct-tag"
)

const StructTag = "url_values"

// Values wraps the standard go library's url.Values to add a Marshal method which uses reflection
// to easily add each unignored field as a query parameter.
type Values struct {
	url.Values
}

func (v *Values) Marshal(values interface{}) (err error) {
	if v.Values == nil {
		v.Values = make(url.Values)
	}

	var errs errors.Errors
	structTag.IterateStructFields(values, func(f reflect.StructField, val reflect.Value) {
		tag := f.Tag.Get(StructTag)
		parts := strings.Split(tag, ",")

		switch parts[0] {
		case "-", "ignore", "ignored":
			return
		}

		field := parts[0]

		var omitEmpty bool
		for _, p := range parts[1:] {
			switch p {
			case "omitempty":
				omitEmpty = true
			}
		}

		if f.Type.Kind() == reflect.Struct {
			return
		}

		data := val.Interface()

		if f.Type.Kind() == reflect.Slice || f.Type.Kind() == reflect.Array {
			// Make sure that the kind here is *not* an array or a struct
			ind := val.Elem()
			for {
				switch ind.Kind() {
				case reflect.Ptr:
					ind = ind.Elem()
				case reflect.Slice:
					return
				case reflect.Struct:
					return
				case reflect.Array:
					return
				}
				break
			}

			// handle it being an array
			darr, ok := data.([]interface{})
			if !ok {
				return
			}
			for _, d := range darr {
				v.Add(field, fmt.Sprintf("%v", d))
			}
			return
		}

		if omitEmpty && val.IsZero() {
			return
		}
		v.Add(field, fmt.Sprintf("%v", data))
	})
	if len(errs) == 0 {
		return nil
	}
	return errs
}
