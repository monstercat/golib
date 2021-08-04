package dbutil

import "github.com/jmoiron/sqlx"

type TxFunc func(tx *sqlx.Tx) error
type ExtFunc func(tx sqlx.Ext) error

func TxNow(db *sqlx.DB, fn TxFunc) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func TxNowMulti(db *sqlx.DB, fn []TxFunc) error {
	return TxNow(db, func(tx *sqlx.Tx) error {
		return RunMulti(tx, fn)
	})
}

func QuickExecTx(tx sqlx.Execer, queries []string, arg ...interface{}) error {
	for _, q := range queries {
		_, err := tx.Exec(q, arg...)
		if err != nil {
			return err
		}
	}
	return nil
}

func RunMulti(tx *sqlx.Tx, fns []TxFunc) error {
	for _, f := range fns {
		if err := f(tx); err != nil {
			return err
		}
	}
	return nil
}

func RunMultiExt(ext sqlx.Ext, fns []ExtFunc) error {
	for _, f := range fns {
		if err := f(ext); err != nil {
			return err
		}
	}
	return nil
}
