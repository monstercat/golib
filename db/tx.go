package dbUtil

import "github.com/jmoiron/sqlx"

func TxNow(db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
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

func TxNowMulti(db *sqlx.DB, fn []func(tx *sqlx.Tx) error) error {
	return TxNow(db, func(tx *sqlx.Tx) error {
		for _, f := range fn {
			if err := f(tx); err != nil {
				return err
			}
		}
		return nil
	})
}
