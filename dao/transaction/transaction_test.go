package transaction

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTx(t *testing.T) {
	// A test context.
	ctx := &TestContext{}
	ctx.On("Commit").Return()

	// A transaction that returns nil, should simply not return an error.
	require.NoError(t, Tx(func(tx *Transaction) error {
		// Get the context.
		_ = GetContext[*TestContext](tx, "test", func() *TestContext {
			return ctx
		})

		// Check that tx has 1 context.
		assert.Len(t, tx.Contexts, 1)
		return nil
	}))

	// Check that commit was called
	ctx.AssertExpectations(t)

	ctx = &TestContext{}
	ctx.On("Rollback").Return()
	err := Tx(func(tx *Transaction) error {
		// Get the context.
		_ = GetContext[*TestContext](tx, "test", func() *TestContext {
			return ctx
		})

		// Check that tx has 1 context.
		assert.Len(t, tx.Contexts, 1)
		return errors.New("12345")
	})
	assert.NotNil(t, err)
	ctx.AssertExpectations(t)
}
