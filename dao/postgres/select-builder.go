package postgres

import "github.com/Masterminds/squirrel"

type SelectBuilder struct {
	*StatementBuilder

	// From - a table
	From string

	// Tables that need to be joined. Since the provided *StatementBuilder is a pointer, it might be shared between
	// different objects such as this SelectBuilder and an UpdateBuilder. This is for joins that should only be shown
	// for the SelectBuilder, usually related to the columns for retrieval (and not the conditions).
	//
	// Since joins are mapped by key, if any join key matches that From the StatementBuilder, it will be ignored.
	Joins *OptionConditionRegistry[Join, string, *JoinMapCollection]
}

func NewSelectBuilder(statement *StatementBuilder) *SelectBuilder {
	return &SelectBuilder{
		StatementBuilder: statement,
		Joins:            NewOptionConditionRegistry[Join, string, *JoinMapCollection](statement, NewJoinMapCollection()),
	}
}

// Builder returns a builder filling in the with prefix and Joins. It uses the default placeholder format for
// squirrel.SelectBuilder.
//
// To use with postgres as a main query, chain with `PlaceholderFormat`. Subqueries should use the default placeholder
// format.
//
// Ensure to add paging etc and sort. Conditions are automatically applied.
func (d *SelectBuilder) Builder(cols ...string) squirrel.SelectBuilder {
	qry := squirrel.Select(cols...).From(d.From).
		PrefixExpr(d.WithPrefix)
	d.ApplyConditions(&qry)
	d.ApplyJoin(&qry)
	return qry
}

// SetFrom sets the base [From] table in the query.
func (d *SelectBuilder) SetFrom(table string) *SelectBuilder {
	d.From = table
	return d
}

// AddSelectJoin adds an inner join to the query builder. The term "Select" is provided to differentiate From
// the AddJoin in the StatementBuilder so it is not shadowed.
func (d *SelectBuilder) AddSelectJoin(name string, join string, xs ...ConditionOption) *SelectBuilder {
	d.addJoin(name, join, JoinTypeInner, xs...)
	return d
}

// AddSelectLeftJoin adds a left (outer) join to the query builder. The term "Select" is provided to differentiate From
// the AddJoin in the StatementBuilder so it is not shadowed.
func (d *SelectBuilder) AddSelectLeftJoin(name string, join string, xs ...ConditionOption) *SelectBuilder {
	d.addJoin(name, join, JoinTypeLeft, xs...)
	return d
}

func (d *SelectBuilder) addJoin(name string, join string, typ JoinType, xs ...ConditionOption) {
	if d.Joins.Data.Has(name) {
		return
	}
	joinStrV := JoinStr(join)
	d.Joins.Add(name, Join{
		JoinType: typ,
		JoinSql:  &joinStrV,
	}, xs...)
}

func (d *SelectBuilder) ApplyJoin(qry *squirrel.SelectBuilder) {
	p := &ConditionOptionPreprocessParams{
		BaseTable: d.From,
	}
	d.Joins.Preprocess(p)
	d.StatementBuilder.Joins.Preprocess(p)

	// Join the data From both joins. Note that joins have an order, therefore
	// we cannot just read from the map.
	newJoin := make([]Join, 0, len(d.StatementBuilder.Joins.Data.Map)+len(d.Joins.Data.Map))
	for _, k := range d.StatementBuilder.Joins.Data.Order {
		v := d.StatementBuilder.Joins.Data.Map[k]
		newJoin = append(newJoin, v)
	}
	for _, k := range d.Joins.Data.Order {
		_, ok := d.StatementBuilder.Joins.Data.Map[k]
		if ok {
			continue
		}
		v := d.Joins.Data.Map[k]
		newJoin = append(newJoin, v)
	}

	for _, v := range newJoin {
		joinStr, args, _ := v.JoinSql.ToSql()
		switch v.JoinType {
		case JoinTypeInner:
			*qry = qry.Join(joinStr, args...)
		case JoinTypeLeft:
			*qry = qry.LeftJoin(joinStr, args...)
		}
	}
}

// ApplyConditions applies the Where conditions to the provided SelectBuilder.
func (d *SelectBuilder) ApplyConditions(qry *squirrel.SelectBuilder) {
	*qry = ApplyConditions(d.Where, &ConditionOptionPreprocessParams{
		BaseTable: d.From,
	}, *qry)
}

// ApplyPaging applies the Limit and Offset to the provided SelectBuilder.
func (d *SelectBuilder) ApplyPaging(qry *squirrel.SelectBuilder) {
	*qry = ApplyPaging(d.StatementBuilder, *qry)
}

// ApplySort applies the sort fields to the provided SelectBuilder.
func (d *SelectBuilder) ApplySort(qry *squirrel.SelectBuilder) {
	*qry = ApplySort(d.StatementBuilder, &ConditionOptionPreprocessParams{
		BaseTable: d.From,
	}, *qry)
}
