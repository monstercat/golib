package postgres

import (
	"strings"

	"github.com/Masterminds/squirrel"
)

const TablePlaceholder = "[[table]]"

// ReplaceTablePlaceholderOption instructs the query builder to replace the TablePlaceholder occurrences in
// the SQL with a proper table, defined by the [From].
var ReplaceTablePlaceholderOption QueryBuilderConditionOptionFunc = func(s squirrel.Sqlizer) squirrel.Sqlizer {
	return &ReplaceTablePlaceholder{
		Sqlizer: s,
	}
}

// ReplaceTablePlaceholder is a squirrel.Sqlizer that adds a table to an existing SQL string. It looks for a placeholder [[table]]
// which it replaces with the table provided.
type ReplaceTablePlaceholder struct {
	// The table to prepend
	table string

	// The base sqlizer to modify.
	squirrel.Sqlizer
}

// Preprocess implements ConditionOptionPreprocess
func (t *ReplaceTablePlaceholder) Preprocess(q *StatementBuilder, p *ConditionOptionPreprocessParams) {
	t.table = p.BaseTable
}

// ToSql satisfies squirrel.Sqlizer.
func (t *ReplaceTablePlaceholder) ToSql() (string, []interface{}, error) {
	sql, args, err := t.Sqlizer.ToSql()
	if err != nil {
		return "", nil, err
	}
	sql = strings.Replace(sql, TablePlaceholder, t.table, -1)
	return sql, args, nil
}

func WithTablePlaceholder(col string) string {
	return TablePlaceholder + "." + col
}
