package dbUtil

import (
	"fmt"
	"reflect"
	"strings"

	. "github.com/monstercat/golib/string"
	. "github.com/monstercat/golib/struct-tag"
)

const SelectTagName = "select"
const SelectSetTagName = "select-sets"

type Coalescer func(reflect.Type, *SelectTags) string

var CoalesceFromType Coalescer = DefaultCoalescer

var arrIds = []string{"Array", "[]", "FlatMap"}
func DefaultCoalescer(t reflect.Type, s *SelectTags) string {

	if s.UUID {
		return "'00000000-0000-0000-0000-000000000000'"
	}

	name := t.String()

	switch name {
	case "pgnull.NullString", "sql.NullString", "string":
		return "''"
	case "pgnull.NullInt", "sql.NullInt64", "sql.NullFloat64", "int", "int64", "int32", "float32", "float64":
		return "0"
	case "time.Time", "NullTime":
		return "'0000-00-00T00:00:00Z'"
	case "bool", "sql.NullBool":
		return "false"
	}

	for _, v := range arrIds {
		if strings.Index(name, v) > -1 {
			return "'{}'"
		}
	}

	return "''"
}

func SetCoalescer(c Coalescer) {
	CoalesceFromType = c
}

type SelectTags struct {
	Coalesce bool
	Ignore   bool
	UUID     bool
	Sets     []string
}

func (t *SelectTags) Parse(str string) {
	p := strings.Split(str, ",")

	for _, v := range p {
		switch strings.TrimSpace(v) {
		case "coalesce":
			t.Coalesce = true
		case "uuid":
			t.UUID = true
		case "ignore", "-":
			t.Ignore = true
		}
	}
}

func (t *SelectTags) ParseSets(str string) {
	if str == "" {
		return
	}
	t.Sets = strings.Split(str, ",")
}

func (t *SelectTags) ContainsSet(set string) bool {
	if t.Sets == nil || len(t.Sets) == 0 {
		return false
	}
	for _, s := range t.Sets {
		if s == set {
			return true
		}
	}
	return false
}

func (t *SelectTags) Apply(
	column string,
	prefix string,
	f reflect.Type,
) string {
	if prefix != "" && prefix[len(prefix)-1] != '.' {
		prefix = prefix + "."
	}

	if !t.Coalesce {
		return fmt.Sprintf("%s%s", prefix, column)
	}
	def := CoalesceFromType(f, t)
	return fmt.Sprintf("COALESCE(%s%s, %s) as %[2]s", prefix, column, def)
}

func GetColumnsList(val interface{}, prefix string, filterFields ...string) []string {
	return GetColumnsForSet("", val, prefix, filterFields...)
}
func GetColumnsListExcl(val interface{}, prefix string, filterFields ...string) []string {
	return GetColumnsForSetExcl("", val, prefix, filterFields...)
}

func GetColumnsForSet(set string, val interface{}, prefix string, filterFields ...string) []string {
	return getColumnsForSet(set, val, prefix, false, filterFields...)
}
func GetColumnsForSetExcl(set string, val interface{}, prefix string, filterFields ...string) []string {
	return getColumnsForSet(set, val, prefix, true, filterFields...)
}

func getColumnsForSet(set string, val interface{}, prefix string, invert bool, filterFields ...string) []string {
	m := make([]string, 0)
	IterateStructFields(val, func(f reflect.StructField, v reflect.Value) {
		t := &SelectTags{}
		t.Parse(f.Tag.Get(SelectTagName))
		t.ParseSets(f.Tag.Get(SelectSetTagName))

		if !shouldReturnField(filterFields, f.Name, invert) {
			return
		}
		if set != "" && !t.ContainsSet(set) {
			return
		}

		name := ColumnNameFromDbTag(f)
		m = append(m, t.Apply(name, prefix, f.Type))
	})
	return m
}

func GetColumnsByTag(val interface{}, prefix string, filterFields ...string) map[string]string {
	return getColumnsByTag(val, prefix, false, filterFields...)
}
func GetColumnsByTagExcl(val interface{}, prefix string, filterFields ...string) map[string]string {
	return getColumnsByTag(val, prefix, true, filterFields...)
}

func getColumnsByTag(val interface{}, prefix string, invert bool, filterFields ...string) map[string]string {
	m := make(map[string]string)
	IterateStructFields(val, func(f reflect.StructField, v reflect.Value) {
		t := &SelectTags{}
		t.Parse(f.Tag.Get(SelectTagName))

		if !shouldReturnField(filterFields, f.Name, invert) {
			return
		}
		name := ColumnNameFromDbTag(f)
		m[f.Name] = t.Apply(name, prefix, f.Type)
	})

	return m
}

func shouldReturnField(filterFields []string, field string, invert bool) bool {
	// If invert = true, it means exclude those in the list. Therefore,
	// it should return true if StringInList is false.
	// Otherwise, it means include.
	return StringInList(filterFields, field) == !invert
}
