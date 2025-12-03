package transaction

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTx_Rollback(t *testing.T) {
	ctx := &TestTransactionContext{}

	var transaction *Transaction
	err := Tx(func(tx *Transaction) error {
		// Set the context
		tx.Contexts["Test"] = ctx
		transaction = tx
		return errors.New("an error. should rollback")
	})

	// The error should be returned.
	assert.Error(t, err)

	// Transaction should be not nil
	assert.NotNil(t, transaction)

	// It should rollback and not commit
	assert.True(t, ctx.DidRollback)
	assert.False(t, ctx.DidCommit)
}

func TestTx_Commit(t *testing.T) {
	ctx := &TestTransactionContext{}

	var transaction *Transaction
	err := Tx(func(tx *Transaction) error {
		// Set the context
		tx.Contexts["Test"] = ctx
		transaction = tx
		return nil
	})

	// No error should be returned
	assert.Nil(t, err)

	// Transaction should be not nil
	assert.NotNil(t, transaction)

	// It should roll back and not commit
	assert.False(t, ctx.DidRollback)
	assert.True(t, ctx.DidCommit)
}

func TestTx_Panic(t *testing.T) {
	t.Parallel()

	t.Run("Panic!", func(t *testing.T) {
		ctx := &TestTransactionContext{}
		defer func() {
			_ = recover()

			// ctx should have gotten a rollback call.
			assert.True(t, ctx.DidRollback)
			assert.False(t, ctx.DidCommit)
		}()
		_ = Tx(func(tx *Transaction) error {
			// Add the test transaction.
			tx.Contexts["Test"] = ctx

			// Panic!
			panic("testing panic")
		})
	})

	t.Run("Normal", func(t *testing.T) {
		ctx := &TestTransactionContext{}
		_ = Tx(func(tx *Transaction) error {
			// Add the test transaction.
			tx.Contexts["Test"] = ctx
			return nil
		})

		// should not have rolled back.
		assert.False(t, ctx.DidRollback)
		assert.True(t, ctx.DidCommit)
	})
}
