package postgres

import (
	"errors"
	"fmt"

	"github.com/lib/pq"
)

// TreatNullConstraint modifies a PostgreSQL error to include a derived
// constraint name for null constraint violations if not specified or missing.
func TreatNullConstraint(err error) error {
	var e *pq.Error
	if errors.As(err, &e) &&
		string(e.Code) == "23502" {
		// if Contstraint isn't defined explictly in the lm postgres,
		// it will generate the constraint to help pgErrorMap to identify
		// the exact column that violated NOT NULL constraint.
		if e.Constraint == "" {
			e.Constraint = fmt.Sprintf("%s_%s_not_null", e.Table, e.Column)
		}
	}
	return err
}
