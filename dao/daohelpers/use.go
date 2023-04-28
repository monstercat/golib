package daohelpers

import "github.com/monstercat/golib/dao/transaction"

type Using[R any] interface {
	Use(tx *transaction.Transaction) R
}
