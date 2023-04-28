package postgres

import (
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyPaging(t *testing.T) {
	d := NewStatementBuilder()
	d.Limit = 100
	d.Offset = 200

	qry := squirrel.Select("*").From("table")
	qry = ApplyPaging(d, qry)

	sql, _, err := qry.ToSql()
	require.NoError(t, err, "Query could not be generated")
	assert.Contains(t, sql, "OFFSET 200")
	assert.Contains(t, sql, "LIMIT 100")
}
