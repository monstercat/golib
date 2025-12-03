package transaction

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/jmoiron/sqlx"
)

// SqlxConnProvider provides the service details for postgres based services.
type SqlxConnProvider[R any] struct {
	returnVar R
	DB        *sqlx.DB
	ctx       *SqlxTransactionContext

	// The name to use for the context. This should be
	// TransactionKeyPostgres or another key using sqlx
	contextName string
}

func NewPostgresConnProvider[R any](r R, db *sqlx.DB) SqlxConnProvider[R] {
	return SqlxConnProvider[R]{
		returnVar:   r,
		DB:          db,
		contextName: TransactionKeyPostgres,
	}
}

// Return returns the variable that is returned for chaining
func (s *SqlxConnProvider[R]) Return() R {
	return s.returnVar
}

// InTransaction returns whether we are currently in a transaction
func (s *SqlxConnProvider[R]) InTransaction() bool {
	return s.ctx != nil && !s.ctx.isClosed
}

// MustTx starts an independent transaction. It returns the orignial transaction if already in a transaction.
func (s *SqlxConnProvider[R]) MustTx(fn func(tx sqlx.Ext) error) error {
	if s.InTransaction() {
		return fn(s.ctx.Tx)
	}
	return Tx(func(tx *Transaction) error {
		if err := s.UseWithError(tx); err != nil {
			return err
		}
		return fn(s.ctx.Tx)
	})
}

// GetDb returns the appropriate Db. The assumption is that EXT is only not-nil when in a transaction.
func (s *SqlxConnProvider[R]) GetDb() sqlx.Ext {
	if s.ctx != nil && !s.ctx.isClosed {
		return s.ctx.Tx
	}
	return s.DB
}

// UseWithError sets the service to be a postgres service. Also provides a new postgres service. This version of use
// returns an error.
func (s *SqlxConnProvider[R]) UseWithError(t *Transaction) error {
	if t == nil {
		return nil
	}
	if s.contextName == "" {
		s.contextName = TransactionKeyPostgres
	}
	ptctx, err := GetSqlxTransactionContext(s.DB, t, s.contextName)
	if err != nil {
		return err
	}
	s.ctx = ptctx
	return nil
}

// Use sets the service to be a postgres service. This version *ignores* the error.
func (s *SqlxConnProvider[R]) Use(t *Transaction) R {
	s.UseWithError(t)
	return s.returnVar
}

// PostgresJSONB wraps any object to satisfy the Valuer interface for PostgreSQL
// storage.
type PostgresJSONB struct {
	B interface{}
}

// Value satisfies the Valuer interface.
func (p PostgresJSONB) Value() (driver.Value, error) {
	j, err := json.Marshal(p.B)
	return j, err
}

// Scan satisfies the Scanner interface
func (p *PostgresJSONB) Scan(val any) error {
	if val == nil {
		return nil
	}
	byt, ok := val.([]byte)
	if !ok {
		return errors.New("expecting a byte array")
	}
	return json.Unmarshal(byt, p.B)
}
