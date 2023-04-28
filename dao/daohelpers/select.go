package daohelpers

// Selecting provides a quick way to define functions that select one or more
// items from a list of items. Also includes other related functionality such as
// Total, Iterate, and Exists.
//
// This is meant to be used a helper in defining Dao.
type Selecting[R any] interface {
	// Exists returns true if the query returns any results.
	Exists() (bool, error)

	// Total returns the total # of items (ignoring paging)
	Total() (uint64, error)

	// Get returns a single item
	Get() (R, error)

	// Select returns a list of items
	Select() ([]R, error)

	// Iterate returns an object that iterates through a list of results one
	// at a time.
	Iterate() (ObjectIterator[R], error)
}

// ObjectIterator is the return type for Iterate
type ObjectIterator[R any] interface {
	Next() (R, error)
}
