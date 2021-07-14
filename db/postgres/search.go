package pgUtils

import (
	"github.com/Masterminds/squirrel"

	"github.com/monstercat/golib/operator"
)

// configuration for different searches.
type ISearchOperatorConfig interface {
	Apply(o, rem []operator.Operator, a *Accumulator, prefix string)
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
	sqlizer := a.GetCondition()
	if sqlizer != nil {
		*query = query.Where(sqlizer)
	}
}
func (a *Accumulator) GetCondition() squirrel.Sqlizer {
	if len(a.and) > 0 {
		a.or = append(a.or, a.and)
	}
	if len(a.remainders) > 0 {
		a.or = append(a.or, a.remainders)
	}
	if len(a.or) == 0 {
		return nil
	}
	return a.or
}

func ApplyOperators(query *squirrel.SelectBuilder, config []ISearchOperatorConfig, ops operator.Operators, prefix string) {
	a := &Accumulator{}
	for _, c := range config {
		os := ops.Get(c.GetKeys()...)
		c.Apply(os, ops.Remainders, a, prefix)
	}
	a.ApplyToQuery(query)
}

func loopOperators(os []operator.Operator, a *Accumulator, sq func(o operator.Operator) squirrel.Sqlizer) {
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

type SearchOperatorConfigBase struct {
	Keys  []string
	Field string
}

func (c SearchOperatorConfigBase) GetKeys() []string {
	return c.Keys
}
