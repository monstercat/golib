package pgUtils

import (
	"fmt"

	"github.com/Masterminds/squirrel"

	"github.com/monstercat/golib/operator"
)

// configuration for different searches.
type ISearchOperatorConfig interface {
	Apply(o, rem []operator.Operator, a *Accumulator, prefix string)
	IsStringType() bool
	GetKeys() []string
}

type Accumulator struct {
	or         squirrel.Or
	and        squirrel.And
	remainders squirrel.Or
}
func (a *Accumulator) ApplyAnd(sql squirrel.Sqlizer) {
	a.and = append(a.and, sql)
}
func (a *Accumulator) ApplyOr(sql squirrel.Sqlizer) {
	a.or = append(a.or, sql)
}
func (a *Accumulator) ApplyRemainder(sql squirrel.Sqlizer) {
	a.remainders = append(a.remainders, sql)
}
func (a *Accumulator) ApplyToQuery(query *squirrel.SelectBuilder) {
	// TODO: check if this adds the proper brackets
	if len(a.and) > 0 {
		a.or = append(a.or, a.and)
	}
	if len(a.remainders) > 0 {
		a.or = append(a.or, a.remainders)
	}
	if len(a.or) > 0 {
		*query = query.Where(a.or)
	}
}

func ApplyOperators(query *squirrel.SelectBuilder, config []ISearchOperatorConfig, ops operator.Operators, prefix string) {
	a := &Accumulator{}
	for _, c := range config {
		os := ops.Get(c.GetKeys()...)
		c.Apply(os, ops.Remainders, a, prefix)
	}
	a.ApplyToQuery(query)
}



type SearchOperatorConfigBase struct {
	Keys  []string
	Field string
}
func (c SearchOperatorConfigBase) GetKeys() []string {
	return c.Keys
}

type StringSearchOperatorConfigBase struct {
	SearchOperatorConfigBase
}
func (c StringSearchOperatorConfigBase) IsStringType() bool {
	return true
}

type SearchOperatorConfigUUID struct{
	StringSearchOperatorConfigBase
}
func (c SearchOperatorConfigUUID) Apply(os, rem []operator.Operator, a *Accumulator, prefix string) {
	loopOperators(os, a, func(o operator.Operator) squirrel.Sqlizer {
		return squirrel.Expr(fmt.Sprintf("%s%s%s = ?", prefix, c.Field, doNot(o)), o.Value)
	})
}
func NewUUIDOperator(field string, keys ...string) SearchOperatorConfigUUID {
	return SearchOperatorConfigUUID{
		StringSearchOperatorConfigBase{
			SearchOperatorConfigBase{
				Field: field, Keys: keys,
			},
		},
	}
}

type SearchOperatorConfigStringLike struct {
	StringSearchOperatorConfigBase
}
func (c SearchOperatorConfigStringLike) Sqlizer(o operator.Operator, prefix string) squirrel.Sqlizer {
	return squirrel.Expr(fmt.Sprintf("%s%s%s ILIKE ?", prefix, c.Field, doNot(o)), "%"+o.Value+"%")
}
func (c SearchOperatorConfigStringLike) Apply(os, rem []operator.Operator, a *Accumulator, prefix string) {
	fn := func(o operator.Operator) squirrel.Sqlizer {
		return c.Sqlizer(o, prefix)
	}
	loopOperators(os, a, fn)
	applyRemainders(rem, a, fn)
}
func NewStringLikeOperator(field string, keys ...string) SearchOperatorConfigStringLike {
	return SearchOperatorConfigStringLike{
		StringSearchOperatorConfigBase{
			SearchOperatorConfigBase{
				Field: field, Keys: keys,
			},
		},
	}
}

func loopOperators(os []operator.Operator, a *Accumulator, sq func(o operator.Operator) squirrel.Sqlizer ) {
	for _, o := range os {
		q := sq(o)
		if o.Has(operator.ModifierOr) {
			a.ApplyOr(q)
		} else {
			a.ApplyAnd(q)
		}
	}
}

func applyRemainders(rem []operator.Operator, a *Accumulator, sq func(o operator.Operator) squirrel.Sqlizer) {
	for _, o := range rem {
		a.ApplyRemainder(sq(o))
	}
}

func doNot(o operator.Operator) string {
	if o.Has(operator.ModifierNot) {
		return " NOT"
	}
	return ""
}
