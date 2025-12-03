package daohelpers

import "github.com/monstercat/golib/transaction"

type Using[R any] interface {
	Use(tx *transaction.Transaction) R
}
