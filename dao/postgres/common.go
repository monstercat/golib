package postgres

import "github.com/Masterminds/squirrel"

// PageApplicable is an interface that defines two methods - Limit and Offset which return a specific type. Together
// with ApplyPaging, defines a type that returns itself.
type PageApplicable[T any] interface {
	Limit(uint64) T
	Offset(uint64) T
}

// ApplyPaging applies the Limit and Offset to the provided Builder.
func ApplyPaging[T PageApplicable[T]](d *StatementBuilder, qry T) T {
	if d.Limit > 0 {
		qry = qry.Limit(d.Limit)
	}
	if d.Offset > 0 {
		qry = qry.Offset(d.Offset)
	}
	return qry
}

// SortApplicable is an interface that defines an OrderBy method which returns a specific type. Together with ApplySort,
// defines a type that returns itself.
type SortApplicable[T any] interface {
	OrderBy(...string) T
}

// ApplySort applies the OrderBy to the provided Builder.
func ApplySort[T SortApplicable[T]](
	d *StatementBuilder,
	p *ConditionOptionPreprocessParams,
	qry T,
) T {
	if d.OrderBys.Data.Size() == 0 {
		return qry
	}
	d.OrderBys.Preprocess(p)
	return qry.OrderBy(d.OrderBys.Data.ToStrings()...)
}

// ConditionApplicable is an interface that defines a Where method which returns a specific type. Together with
// ApplyConditions, defines a type that returns itself.
type ConditionApplicable[T any] interface {
	Where(pred interface{}, args ...interface{}) T
}

// ApplyConditions applies the Where to the provided Builder.
func ApplyConditions[T ConditionApplicable[T]](
	d *SliceOptionConditionRegistry[squirrel.Sqlizer, *SqlizerSliceCollection],
	p *ConditionOptionPreprocessParams,
	qry T,
) T {
	if d.Data.Size() == 0 {
		return qry
	}
	d.Preprocess(p)
	return qry.Where(squirrel.And(d.Data.Slice))
}
