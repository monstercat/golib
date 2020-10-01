package dbUtil

import (
	"reflect"
	"strings"

	"github.com/lib/pq"

	. "github.com/monstercat/golib/struct-tag"
)

// Setmap changes a struct into a
// map ready for database insert or update.
//
// It tries to use the db:"-" and camelcase to
// infer a database field name.

const SetMapTagName = "setmap"

type SetMapTags struct {
	OmitEmpty bool
	OmitInsert bool
	Ignore     bool
}

func ParseSetMapTags(str string) *SetMapTags {
	s := &SetMapTags{}
	s.Parse(str)
	return s
}

func (t *SetMapTags) Parse(str string) {
	p := strings.Split(str, ",")

	for _, v := range p {
		switch strings.TrimSpace(v) {
		case "-", "ignore":
			t.Ignore = true
		case "omitempty":
			t.OmitEmpty = true
		case "omitinsert":
			t.OmitInsert = true
		}
	}
}

type SetMapIterator func(string, interface{})

func IteratePgFields(val interface{}, isInsert bool, it SetMapIterator) {
	IterateStructFields(val, extractSetMapTags(isInsert, it))
}

func extractSetMapTags(isInsert bool, it SetMapIterator) StructFieldIterator {
	return func(f reflect.StructField, v reflect.Value) {
		tag := ParseSetMapTags(f.Tag.Get(SetMapTagName))
		if tag.Ignore {
			return
		}
		if isInsert && tag.OmitInsert {
			return
		}

		data, isZero, checked := GetData(v)
		if !checked {
			if f.Type.Kind() == reflect.Struct {
				// This function doesn't handle structs other
				// than the ones specifically defined above.
				return
			}
			if f.Type.Kind() == reflect.Array || f.Type.Kind() == reflect.Slice {
				// If data is an array, we need to wrap it as a
				// postgresable array.
				data = pq.Array(data)
				isZero = false
			}
		}

		if isZero && tag.OmitEmpty {
			return
		}

		name := ColumnNameFromDbTag(f)
		if name == "" {
			return
		}
		it(name, data)
	}
}

func SetMap(val interface{}, isInsert bool) map[string]interface{} {
	m := map[string]interface{}{}
	IteratePgFields(val, isInsert, func(col string, data interface{}) {
		m[col] = data
	})
	return m
}

