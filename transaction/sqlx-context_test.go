package transaction

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgresTransactionContext_Rollback(t *testing.T) {
	_db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	db := sqlx.NewDb(_db, "sqlmock")
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	err := Tx(func(tx *Transaction) error {
		// Get should also create a context. By creating, a transaction
		// begin should be expected. Furthermore, the context should be
		// directly added into the transaction.
		//
		// We don't need to check this. If it isn't properly added,
		// rollback will *not* occur.
		_, err := GetSqlxTransactionContext(db, tx, TransactionKeyPostgres)
		if err != nil {
			return err
		}
		return errors.New("error. should rollback")
	})

	assert.Error(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
}

func TestPostgresTransactionContext_Commit(t *testing.T) {
	_db, mock, _ := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	db := sqlx.NewDb(_db, "sqlmock")
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	err := Tx(func(tx *Transaction) error {
		// Get should also create a context. By creating, a transaction
		// begin should be expected. Furthermore, the context should be
		// directly added into the transaction.
		//
		// We don't need to check this. If it isn't properly added,
		// commit will *not* occur.
		_, err := GetSqlxTransactionContext(db, tx, TransactionKeyPostgres)
		if err != nil {
			return err
		}
		return nil
	})

	assert.Nil(t, err)
	assert.Nil(t, mock.ExpectationsWereMet())
}
