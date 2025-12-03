package transaction

import "sync"

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
	rw sync.RWMutex

	// Contexts are what is used to retrieve any objects required to be stored within the
	Contexts map[string]Context

	// Whether the transaction has been closed or not.
	isClosed bool
}

// TransactionContextNumQueries defines a transaction context that can return the
// number of queries that have been executed. It is an optional method for
// trnascation contexts.
type TransactionContextNumQueries interface {
	// NumQueries returns the number of queries that have been executed
	// within the transaction
	NumQueries() int
}

// NumQueries returns the number of queries that have been run in all
// transaction contexts, for those that support it.
func (t *Transaction) NumQueries() int {
	total := 0
	for _, c := range t.Contexts {
		if n, ok := c.(TransactionContextNumQueries); ok {
			total += n.NumQueries()
		}
	}
	return total
}

func (t *Transaction) IsClosed() bool {
	t.rw.RLock()
	defer t.rw.RUnlock()

	return t.isClosed
}

func (t *Transaction) Close() {
	t.rw.Lock()
	defer t.rw.Unlock()

	t.isClosed = true
}

func (t *Transaction) Lock() {
	t.rw.Lock()
}

func (t *Transaction) Unlock() {
	t.rw.Unlock()
}

func (t *Transaction) SetContext(key string, ctx Context) {
	t.rw.Lock()
	defer t.rw.Unlock()

	t.Contexts[key] = ctx
}

func (t *Transaction) HasContext(key string) (Context, bool) {
	t.rw.RLock()
	defer t.rw.RUnlock()

	ctx, ok := t.Contexts[key]
	return ctx, ok
}

// Rollback the transaction. Any changes made will be ignored. Once called,
// the transaction is deemed closed. On a closed transaction, does nothing.
func (t *Transaction) Rollback() {
	if t.IsClosed() {
		return
	}
	t.Close()
	for _, c := range t.Contexts {
		c.Rollback()
	}
}

// Commit the transaction. Changes will be made permanent. Once called,
// the transaction is deemed closed. On a closed transaction, does nothing.
func (t *Transaction) Commit() {
	if t.IsClosed() {
		return
	}
	t.Close()
	for _, c := range t.Contexts {
		c.Commit()
	}
}

// Tx provides the transaction for wrapping of functions. Functions that wish to sue the function should take in the
// Transaction that is passed as the argument.
func Tx(fn func(tx *Transaction) error) error {
	tx := New()

	// Rollback is required here in case of panic.
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// New creates a new transaction object.
func New() *Transaction {
	return &Transaction{
		Contexts: make(map[string]Context),
	}
}

// MustTx ensures that the tx provided to the function is non-nil.
func MustTx(tx *Transaction, fn func(tx *Transaction) error) error {
	if tx != nil {
		return fn(tx)
	}
	return Tx(fn)
}
