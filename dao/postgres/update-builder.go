package postgres

import (
	"fmt"

	"github.com/Masterminds/squirrel"
)

// UpdateBuilder generates an update method based on the encapsulated statement builder.
type UpdateBuilder struct {
	// StatementBuilder is used to provide the conditions for the builder. This is a pointer so that the same instance
	// can be used for both the UpdateBuilder and the SelectBuilder.
	*StatementBuilder

	// Update [table]
	table string

	// Name of the id column (for use with complicated WHERE clauses)
	IdColumnName string
}

func NewUpdateBuilder(statement *StatementBuilder) *UpdateBuilder {
	return &UpdateBuilder{
		StatementBuilder: statement,
	}
}

// SetBaseTable sets the base table in the query.
func (d *UpdateBuilder) SetBaseTable(table string) *UpdateBuilder {
	d.table = table
	return d
}

func (d *UpdateBuilder) generateConditionOptionPreprocessParams() *ConditionOptionPreprocessParams {
	return &ConditionOptionPreprocessParams{
		BaseTable: d.table,
	}
}

func (d *UpdateBuilder) generateAdditionalWith() squirrel.SelectBuilder {
	stmt := &StatementBuilder{
		Joins:    d.Joins,
		Where:    d.Where,
		OrderBys: d.OrderBys,
		Limit:    d.Limit,
		Offset:   d.Offset,
	}
	return NewSelectBuilder(stmt).
		SetFrom(d.table).
		Builder(d.IdColumnName)
}

func (d *UpdateBuilder) preparePrefixExprForCondition() (With, squirrel.Sqlizer) {
	with := make(With)
	for k, v := range d.WithPrefix {
		with[k] = v
	}
	with[d.table+"_condition"] = d.generateAdditionalWith()
	return with, squirrel.Expr(fmt.Sprintf("%s IN (SELECT %[1]s FROM %[2]s_condition)", d.IdColumnName, d.table))
}

// UpdateBuilder returns a builder filling in the with prefix. If joins are required, an additional WITH will be created
// using the base table, joined with all the other tables, and containing the Where clause.
func (d *UpdateBuilder) UpdateBuilder(setMap map[string]interface{}) squirrel.UpdateBuilder {
	qry := squirrel.Update(d.table).SetMap(setMap)
	if d.HasJoins() {
		with, where := d.preparePrefixExprForCondition()
		qry = qry.PrefixExpr(with).Where(where)
	} else {
		params := d.generateConditionOptionPreprocessParams()
		qry = ApplyPaging(d.StatementBuilder, qry)
		qry = ApplySort(d.StatementBuilder, params, qry)
		qry = ApplyConditions(d.Where, params, qry)
	}
	return qry
}

// InsertBuilder returns a builder with the provided set information and the predefined table.
func (d *UpdateBuilder) InsertBuilder(setMap map[string]interface{}) squirrel.InsertBuilder {
	// Squirrel.InsertBuilder *requires* insert queries to have data. However,
	// it is possible to have insert queries with only DEFAULT VALUES.
	return squirrel.Insert(d.table).SetMap(setMap)
}

func (d *UpdateBuilder) DeleteBuilder() squirrel.DeleteBuilder {
	qry := squirrel.Delete(d.table)
	if d.HasJoins() {
		with, where := d.preparePrefixExprForCondition()
		qry = qry.PrefixExpr(with).Where(where)
	} else {
		params := d.generateConditionOptionPreprocessParams()
		qry = ApplyPaging(d.StatementBuilder, qry)
		qry = ApplySort(d.StatementBuilder, params, qry)
		qry = ApplyConditions(d.Where, params, qry)
	}
	return qry
}
