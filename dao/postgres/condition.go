package postgres

import (
	"github.com/Masterminds/squirrel"
	"github.com/cyc-ttn/go-collections"
)

// SqlizerSliceCollection is a slice of Sqlizer which implements Collection
type SqlizerSliceCollection struct {
	SliceCollection[squirrel.Sqlizer]
}

// Apply adds the condition to the item.
func (s *SqlizerSliceCollection) Apply(o ConditionOption, item squirrel.Sqlizer) (squirrel.Sqlizer, squirrel.Sqlizer) {
	modded := o.ModifyCondition(item)
	return modded, modded
}

// ToStrings converts the sqlizer items to strings.
func (s *SqlizerSliceCollection) ToStrings() []string {
	return collections.Map[string, squirrel.Sqlizer](s.Slice, func(agg []string, s squirrel.Sqlizer) (string, bool) {
		sql, _, _ := s.ToSql()
		return sql, true
	})
}
