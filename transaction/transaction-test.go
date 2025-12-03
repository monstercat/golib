package transaction

// TestTransactionContext is a struct that can be used for testing transaction
// contexts.
type TestTransactionContext struct {
	// Was the context committed?
	DidCommit bool

	// Was it rollled back?
	DidRollback bool
}

// Rollback are operations that need to be done to clean-up on failure.
func (t *TestTransactionContext) Rollback() {
	t.DidRollback = true
}

// Commit solidifies the resultant operations.
func (t *TestTransactionContext) Commit() {
	t.DidCommit = true
}
