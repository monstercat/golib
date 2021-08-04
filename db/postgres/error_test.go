package pgutil

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/lib/pq"
)

func TestTransformError(t *testing.T) {

	m := map[ErrorConfig]error{
		MatchesConstraint("uniq_123"):  errors.New("unique_123"),
		MatchesConstraint("help_456"):  errors.New("help_456"),
		MatchesRoutine("some_routine"): errors.New("some_routine"),
	}

	if TransformError(nil, m) != nil {
		t.Error("Expecting nil to return as nil")
	}

	if TransformError(sql.ErrNoRows, m) != sql.ErrNoRows {
		t.Error("Expecting non-pq errors to be returned as is")
	}

	for c, err := range m {
		var perr *pq.Error
		switch v := c.(type) {
		case MatchesConstraint:
			perr = &pq.Error{
				Constraint: string(v),
			}
		case MatchesRoutine:
			perr = &pq.Error{
				Routine: string(v),
			}
		}
		if perr == nil {
			continue
		}

		rerr := TransformError(perr, m)
		if rerr != err {
			t.Errorf("Expecting contraint %s to return error %s, but got %s", c, err, rerr)
		}
	}
}
