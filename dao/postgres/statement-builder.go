package postgres

import (
	"github.com/Masterminds/squirrel"
	"github.com/cyc-ttn/go-collections"
	"github.com/lib/pq"
)

// StatementBuilder helps with building select, update and delete queries.
type StatementBuilder struct {
	// Tables to add to the top. Use the map to make sure that if it has been added or not.
	WithPrefix With

	// Tables to be joined. The reason that this is part of the statement builder is because sometimes the conditions
	// require that the table be included.
	//
	// Joins only make sense in the context of SelectBuilder. For UpdateBuilder, joins will be merged as the last
	// WithPrefix statement as a sub-select, so that it can be used in the WHERE statement properly.
	Joins *OptionConditionRegistry[Join, string, *JoinMapCollection]

	// Where conditions for postgres querying.
	Where *SliceOptionConditionRegistry[squirrel.Sqlizer, *SqlizerSliceCollection]

	// OrderBys are the order-by column
	OrderBys *SliceOptionConditionRegistry[squirrel.Sqlizer, *SqlizerSliceCollection]

	// Limit is the Limit for the query
	Limit uint64

	// Offset is an Offset for the query.
	Offset uint64
}

// NewStatementBuilder generates a query builder and initiates all required internal variables.
func NewStatementBuilder() *StatementBuilder {
	p := &StatementBuilder{}
	p.ClearConditions()
	p.Joins = NewOptionConditionRegistry[Join, string, *JoinMapCollection](p, NewJoinMapCollection())
	p.OrderBys = NewSliceConditionRegistry[squirrel.Sqlizer, *SqlizerSliceCollection](p, &SqlizerSliceCollection{})
	return p
}

func (d *StatementBuilder) ClearConditions() {
	d.Where = NewSliceConditionRegistry[squirrel.Sqlizer, *SqlizerSliceCollection](d, &SqlizerSliceCollection{})
}

// HasJoins returns true if joins are present.
func (d *StatementBuilder) HasJoins() bool {
	return d.Joins.Data.Size() > 0
}

// HasConditions returns true if conditions are present.
func (d *StatementBuilder) HasConditions() bool {
	return d.Where.Data.Size() > 0
}

// GetLimit returns the Limit Data on the StatementBuilder
func (d *StatementBuilder) GetLimit() uint64 {
	return d.Limit
}

// GetOffset returns the Limit Data on the StatementBuilder
func (d *StatementBuilder) GetOffset() uint64 {
	return d.Offset
}

// AddPrefix adds a prefix to the StatementBuilder.
func (d *StatementBuilder) AddPrefix(name string, prefix squirrel.Sqlizer) *StatementBuilder {
	if d.WithPrefix == nil {
		d.WithPrefix = make(map[string]squirrel.Sqlizer)
	}
	if _, ok := d.WithPrefix[name]; ok {
		return d
	}
	d.WithPrefix[name] = prefix
	return d
}

// AddJoin adds an inner join to the query builder.
func (d *StatementBuilder) AddJoin(name string, join string, xs ...ConditionOption) *StatementBuilder {
	d.addJoin(name, join, JoinTypeInner, xs...)
	return d
}

// AddLeftJoin adds a left (outer) join to the query builder.
func (d *StatementBuilder) AddLeftJoin(name string, join string, xs ...ConditionOption) *StatementBuilder {
	d.addJoin(name, join, JoinTypeLeft, xs...)
	return d
}

func (d *StatementBuilder) addJoin(name string, join string, typ JoinType, xs ...ConditionOption) {
	if d.Joins.Data.Has(name) {
		return
	}
	joinStrV := JoinStr(join)
	d.Joins.Add(name, Join{
		JoinType: typ,
		JoinSql:  &joinStrV,
	}, xs...)
}

// AddCondition adds a "Where" condition to the query.
func (d *StatementBuilder) AddCondition(where squirrel.Sqlizer, xs ...ConditionOption) *StatementBuilder {
	d.Where.Add(where, xs...)
	return d
}

// AddStringArrayCondition adds a special condition that looks like [column]=ANY(?)
func (d *StatementBuilder) AddStringArrayCondition(column string, value pq.StringArray, xs ...ConditionOption) *StatementBuilder {
	if len(value) == 0 {
		return d
	}
	return d.AddCondition(squirrel.Expr(column+"=ANY(?)", value), xs...)
}

// AddStringArrayCondition is a generic version of StatementBuilder.AddStringArrayCondition which helps convert the
// string array into a pq.StringArray
func AddStringArrayCondition[T ~string](d *StatementBuilder, column string, value []T, xs ...ConditionOption) {
	if len(value) == 0 {
		return
	}
	arr := pq.StringArray(collections.ToStringSliceOf[string](value))
	d.AddStringArrayCondition(column, arr, xs...)
}

// SetLimit sets the Limit for the query
func (d *StatementBuilder) SetLimit(limit uint64) *StatementBuilder {
	d.Limit = limit
	return d
}

// SetOffset sets the Offset for the query.
func (d *StatementBuilder) SetOffset(offset uint64) *StatementBuilder {
	d.Offset = offset
	return d
}

// AddSort adds an orderBy to the OrderBys field for the query.
func (d *StatementBuilder) AddSort(sort []string, xs ...ConditionOption) *StatementBuilder {
	for _, s := range sort {
		d.OrderBys.Add(squirrel.Expr(s), xs...)
	}
	return d
}
