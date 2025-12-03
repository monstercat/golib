package postgres

import "github.com/Masterminds/squirrel"

// SqlizerSliceCollection is a slice of Sqlizer which implements Collection
type SqlizerSliceCollection struct {
	SliceCollection[squirrel.Sqlizer]
}

// Apply adds the condition to the item.
func (s *SqlizerSliceCollection) Apply(o ConditionOption, item squirrel.Sqlizer) (squirrel.Sqlizer, squirrel.Sqlizer) {
	modded := o.ModifyCondition(item)
	return modded, modded
}

// StringSliceCollection is a slice of strings which implements Collection
type StringSliceCollection struct {
	SliceCollection[string]
}

// Apply adds the condition to the item.
func (s *StringSliceCollection) Apply(o ConditionOption, item string) (string, squirrel.Sqlizer) {
	sql := squirrel.Expr(item)
	modded := o.ModifyCondition(sql)

	moddedSql, _, _ := modded.ToSql()
	return moddedSql, modded
}
