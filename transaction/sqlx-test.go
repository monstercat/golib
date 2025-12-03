//go:build test

package transaction

import "github.com/jmoiron/sqlx"

// UseSqlTx directly injects a transaction to use. This is used primarily
// for testing purposes when the *transaction.Transaction is *not* being used.
//
// This function should be removed when all legacy code is removed.
func (s *SqlxConnProvider[R]) UseSqlTx(tx *sqlx.Tx) {
	s.ctx = &SqlxTransactionContext{
		Tx: &txProxy{Tx: tx},
	}
}
