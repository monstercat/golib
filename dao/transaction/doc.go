// Package transaction provides utilities for wrapping operations in a
// transaction. Upon transaction failure, changes are expected to be rolled
// back, and, upon transaction success, changes are committed.
//
// A transaction object contains a list of contexts, each containing their
// own key (string). Keys should be unique for each transaction type. For
// example, given a postgres connection and a file service, a
// TransactionContext can be created for each, satisfying the Context
// interface. This transaction context can be directly registered into the
// Transaction.Context map. Then, define a method such as `Use(tx *Transaction)`
// which attempts to retrieve the context from the tx.Context map. If it doesn't
// exist, create a new one.
//
//	func GetCustomTransactionContext(tx *Transaction, args AnyArgsNeeded) *CustomTransactionContext{
//	   m, ok := tx.Contexts["CustomTransactionContext"]
//	   if ok {
//	       v, ok := m.(*CustomTransactionContext)
//	       if ok {
//	          return v
//	       }
//	   }
//
//	   v := CreateTransactionContext(args AnyArgsNeeded)
//	   tx.Contexts["CustomTransactionContext"] = v
//	   return v
//	}
//
// A helper method GetContext is created to help with this logic.
package transaction
