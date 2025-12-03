package transaction

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

	"github.com/stretchr/testify/assert"
)

// Test the in-transaction function.
func TestPostgresConnProvider_InTransaction(t *testing.T) {
	_db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	db := sqlx.NewDb(_db, "sqlmock")
	defer db.Close()

	prov := NewPostgresConnProvider[any](nil, db)

	// We aren't in a transaction. The database should be the same pointer
	// as the one that is passed in.
	assert.False(t, prov.InTransaction())
	assert.Equal(t, db, prov.GetDb())

	// In a transaction now. (Mock needs to expect begin / commit)
	mock.ExpectBegin()
	mock.ExpectCommit()
	assert.Nil(t, Tx(func(tx *Transaction) error {
		prov.Use(tx)

		// We are currently in a transaction.
		assert.True(t, prov.InTransaction())
		assert.NotNil(t, prov.ctx.Tx)
		assert.Equal(t, prov.ctx.Tx, prov.GetDb())
		return nil
	}))

	// We are now outside the transaction. isClosed should be true. Tx might
	// not be nil.
	assert.True(t, prov.ctx.isClosed)
	assert.False(t, prov.InTransaction())

	// Test if a new transaction object is being created. This time, we are
	// using MustTx.
	oldTx := prov.ctx.Tx

	mock.ExpectBegin()
	mock.ExpectCommit()
	assert.Nil(t, prov.MustTx(func(tx sqlx.Ext) error {
		assert.True(t, prov.InTransaction())
		assert.NotEqual(t, oldTx, prov.ctx.Tx)
		return nil
	}))

	// We are now outside the transaction. isClosed should be true. Tx might
	// not be nil.
	assert.True(t, prov.ctx.isClosed)
	assert.False(t, prov.InTransaction())

	// Finally, we need to test MustTx when inside a transaction already.
	mock.ExpectBegin()
	mock.ExpectCommit()
	assert.Nil(t, Tx(func(tx *Transaction) error {
		prov.Use(tx)

		assert.True(t, prov.InTransaction())
		currentTx := prov.ctx.Tx

		return prov.MustTx(func(tx sqlx.Ext) error {
			assert.True(t, prov.InTransaction())
			assert.Equal(t, prov.GetDb(), prov.ctx.Tx)

			// Ensure the same transaction object is being used.
			assert.Equal(t, prov.GetDb(), currentTx)
			return nil
		})
	}))
}
