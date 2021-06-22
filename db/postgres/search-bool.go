package pgUtils

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"

	"github.com/monstercat/golib/operator"
)

func NewBoolOperator(field string, keys ...string) SearchOperatorConfigBool {
	return SearchOperatorConfigBool{
		SearchOperatorConfigBase{
			Field: field,
			Keys: keys,
		},
	}
}

type SearchOperatorConfigBool struct {
	SearchOperatorConfigBase
}

func parseBool(o string) bool {
	if o == "" {
		return false
	}
	switch strings.ToLower(o) {
	case "t", "true", "1":
		return true
	}
	return false
}

func (c SearchOperatorConfigBool) Apply(os, rem []operator.Operator, a *Accumulator, prefix string) {
	loopOperators(os, a, func(o operator.Operator) squirrel.Sqlizer {
		b := parseBool(o.Value)
		if o.Has(operator.ModifierNot) {
			b = !b
		}
		if b {
			return squirrel.Expr(fmt.Sprintf("%s%s", prefix, c.Field))
		}
		return squirrel.Expr(fmt.Sprintf("NOT %s%s", prefix, c.Field))
	})
}
