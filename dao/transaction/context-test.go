package transaction

import "github.com/stretchr/testify/mock"

// TestContext is a test transaction context. It implement Context.
type TestContext struct {
	mock.Mock
}

// Rollback are operations that need to be done to clean-up on failure.
func (t *TestContext) Rollback() {
	t.Called()
}

// Commit solidifies the resultant operations.
func (t *TestContext) Commit() {
	t.Called()
}
