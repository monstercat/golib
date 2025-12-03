package postgres

import (
	"github.com/lib/pq"
	pgutil "github.com/monstercat/golib/db/postgres"

	"github.com/monstercat/golib/daohelpers"
)

type PgErrorMap map[pgutil.ErrorConfig]error

var (
	ErrorMap = PgErrorMap{
		pgutil.MatchesCode(pgutil.ErrCodeDuplicate): daohelpers.ErrDuplicate,
	}
)

// MatchesNull is a helper for testing if a pq.Error matches a
// column that is null.
type MatchesNull string

func (c MatchesNull) Test(e *pq.Error) bool {
	return e.Code == "23502" && e.Column == string(c)
}

// MatchesMessage is a helper for testing if a pq.Error matches a
// specific message.
type MatchesMessage string

func (c MatchesMessage) Test(e *pq.Error) bool {
	return e.Message == string(c)
}
