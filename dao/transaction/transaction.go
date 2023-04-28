package transaction

// Context defines a group of transactions in the same type of service (e.g., postgers; s3) which can
// Rollback or Commit.
type Context interface {
	// Rollback are operations that need to be done to clean-up on failure.
	Rollback()

	// Commit solidifies the resultant operations.
	Commit()
}

// Transaction defines a single transaction
type Transaction struct {
	// Contexts are what is used to retrieve any objects required to be stored within the
	Contexts map[string]Context
}

func (t *Transaction) Rollback() {
	for _, c := range t.Contexts {
		c.Rollback()
	}
}

func (t *Transaction) Commit() {
	for _, c := range t.Contexts {
		c.Commit()
	}
}

// Tx provides the transaction for wrapping of functions. Functions that wish to sue the function should take in the
// Transaction that is passed as the argument.
func Tx(fn func(tx *Transaction) error) error {
	tx := &Transaction{
		Contexts: make(map[string]Context),
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// GetContext is a helper method to retrieve a context of a certain type from
// a transaction, associated with a specific key. If it does not exist, it will
// be created through the provided create method.
func GetContext[T Context](tx *Transaction, key string, create func() T) T {
	m, ok := tx.Contexts[key]
	if ok {
		v, ok := m.(T)
		if ok {
			return v
		}
	}

	v := create()
	tx.Contexts[key] = v
	return v
}
