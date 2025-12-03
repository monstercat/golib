package transaction

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

const (
	TransactionKeyPostgres = "postgres"
)

// txProxy is our own wrapper around transactions. It allows us to view the # of
// queries that have been run.
type txProxy struct {
	*sqlx.Tx
	count int
}

func (tx *txProxy) Query(query string, args ...interface{}) (*sql.Rows, error) {
	tx.count++
	return tx.Tx.Query(query, args...)
}
func (tx *txProxy) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	tx.count++
	return tx.Tx.Queryx(query, args...)
}
func (tx *txProxy) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	tx.count++
	return tx.Tx.QueryRowx(query, args...)
}
func (tx *txProxy) Exec(query string, args ...interface{}) (sql.Result, error) {
	tx.count++
	return tx.Tx.Exec(query, args...)
}

type SqlxTransactionContext struct {
	isClosed bool
	Tx       *txProxy
}

func (c *SqlxTransactionContext) Close() {
	c.isClosed = true
}

func (c *SqlxTransactionContext) Commit() {
	c.Tx.Commit()
	c.Close()
}

func (c *SqlxTransactionContext) Rollback() {
	c.Tx.Rollback()
	c.Close()
}

// NumQueries returns the number of queries that have been run
// in this transaction.
func (c *SqlxTransactionContext) NumQueries() int {
	return c.Tx.count
}

func GetSqlxTransactionContext(DB *sqlx.DB, t *Transaction, typ string) (*SqlxTransactionContext, error) {
	x, ok := t.HasContext(typ)
	if !ok {
		return CreateSqlxTransactionContext(DB, t, typ)
	}
	v, ok := x.(*SqlxTransactionContext)
	if !ok || v.isClosed {
		return CreateSqlxTransactionContext(DB, t, typ)
	}
	return v, nil
}

func CreateSqlxTransactionContext(DB *sqlx.DB, t *Transaction, typ string) (*SqlxTransactionContext, error) {
	// We want to lock manually so the whole process is locked.
	t.Lock()
	defer t.Unlock()

	tx, err := DB.Beginx()
	if err != nil {
		return nil, err
	}

	n := &SqlxTransactionContext{
		Tx: &txProxy{Tx: tx},
	}
	t.Contexts[typ] = n
	return n, nil
}
