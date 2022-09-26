package pgutil

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"

	"github.com/monstercat/golib/operator"
)

func NewSearchOperatorConfigBase(field string, keys ...string) SearchOperatorConfigBase {
	return SearchOperatorConfigBase{Field: field, Keys: keys}
}

func NewUUIDOperator(field string, keys ...string) SearchOperatorConfigUUID {
	return SearchOperatorConfigUUID{NewSearchOperatorConfigBase(field, keys...)}
}

func NewStringLikeOperator(field string, keys ...string) SearchOperatorConfigStringLike {
	return SearchOperatorConfigStringLike{NewSearchOperatorConfigBase(field, keys...)}
}

type SearchOperatorConfigUUID struct {
	SearchOperatorConfigBase
}

func (c SearchOperatorConfigUUID) Apply(os, rem []operator.Operator, a *Accumulator, prefix string) {
	loopOperators(os, a, func(o operator.Operator) squirrel.Sqlizer {
		return squirrel.Expr(fmt.Sprintf("%s%s%s = ANY(?)", prefix, c.Field, doNot(o)), pq.StringArray(o.Values))
	})
}

type SearchOperatorConfigStringLike struct {
	SearchOperatorConfigBase
}

func (c SearchOperatorConfigStringLike) Sqlizer(o operator.Operator, prefix string) squirrel.Sqlizer {
	and := squirrel.And{}
	for _, v := range o.Values {
		// Escape _
		parts := strings.Split(v, "_")
		value := strings.Join(parts, "\\_")

		// Escape the %
		parts = strings.Split(value, "%")
		value = strings.Join(parts, "\\%")

		and = append(and, squirrel.Expr(fmt.Sprintf("%s%s%s ILIKE ?", prefix, c.Field, doNot(o)), "%"+value+"%"))
	}
	return and
}

func (c SearchOperatorConfigStringLike) Apply(os, rem []operator.Operator, a *Accumulator, prefix string) {
	fn := func(o operator.Operator) squirrel.Sqlizer {
		return c.Sqlizer(o, prefix)
	}
	loopOperators(os, a, fn)
	applyRemainders(rem, a, fn)
}
