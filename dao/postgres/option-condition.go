package postgres

import "github.com/Masterminds/squirrel"

// ConditionOption is an option that modifies a query string sqlizer for a condition.
type ConditionOption interface {
	ModifyCondition(s squirrel.Sqlizer) squirrel.Sqlizer
}

// ConditionOptionPreprocessParams are parameters that can be passed into the params, which are unrelated to conditions.
type ConditionOptionPreprocessParams struct {
	// BaseTable for SELECT is the table after [FROM] keyword. For update and delete, it is the table that is being
	// updated or deleted From (i.e., the table following the [UPDATE] and [DELETE] keywords, correspondingly).
	BaseTable string
}

// ConditionOptionPreprocess denotes that the ConditionOption can be registered for preprocessing
// before it is applied to the query. It does this by passing the whole *StatementBuilder as a preprocessing step.
//
// See QueryCondition.Apply.
type ConditionOptionPreprocess interface {
	Preprocess(q *StatementBuilder, p *ConditionOptionPreprocessParams)
}

// QueryBuilderConditionOptionFunc converts a func to satisfy ConditionOption
type QueryBuilderConditionOptionFunc func(s squirrel.Sqlizer) squirrel.Sqlizer

// ModifyCondition satisfies ConditionOption
func (f QueryBuilderConditionOptionFunc) ModifyCondition(s squirrel.Sqlizer) squirrel.Sqlizer {
	return f(s)
}

// SubQueryOption wraps the string array condition in a subquery. In particular, it adds it as a WHERE condition
// in the provided squirrel.SelectBuilder.
func SubQueryOption(subQ squirrel.SelectBuilder) QueryBuilderConditionOptionFunc {
	return func(s squirrel.Sqlizer) squirrel.Sqlizer {
		return subQ.Where(s)
	}
}
