package pgUtils

import (
	"fmt"

	"github.com/Masterminds/squirrel"

	"github.com/monstercat/golib/operator"
)

func NewStringSearchOperator(field string, keys ...string) SearchOperatorConfigBase {
	return SearchOperatorConfigBase{Field: field, Keys: keys}
}

func NewUUIDOperator(field string, keys ...string) SearchOperatorConfigUUID {
	return SearchOperatorConfigUUID{NewStringSearchOperator(field, keys...)}
}

func NewStringLikeOperator(field string, keys ...string) SearchOperatorConfigStringLike {
	return SearchOperatorConfigStringLike{NewStringSearchOperator(field, keys...)}
}

type SearchOperatorConfigUUID struct {
	SearchOperatorConfigBase
}

func (c SearchOperatorConfigUUID) Apply(os, rem []operator.Operator, a *Accumulator, prefix string) {
	loopOperators(os, a, func(o operator.Operator) squirrel.Sqlizer {
		return squirrel.Expr(fmt.Sprintf("%s%s%s = ?", prefix, c.Field, doNot(o)), o.Value)
	})
}

type SearchOperatorConfigStringLike struct {
	SearchOperatorConfigBase
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

