package pgUtils

import "github.com/lib/pq"

type ErrorConfig interface {
	Test(e *pq.Error) bool
}

type MatchesConstraint string

func (c MatchesConstraint) Test(e *pq.Error) bool {
	return e.Constraint == string(c)
}

// Transforms the error coming in *if* it is a PG error, based on the
// configuration.
//
// Usage:
//  cfg := map[ErrorConfig]error{
//     MatchesConstraint("id_fkey"): ErrInvalidId,
//  }
//  ...
//   _, err := db.Exec(....)
//   return TransformError(err, cfg)
func TransformError(err error, cfg map[ErrorConfig]error) error {
	if err == nil {
		return nil
	}
	perr, ok := err.(*pq.Error)
	if !ok {
		return err
	}
	for c, err := range cfg {
		if c.Test(perr) {
			return err
		}
	}
	return err
}
