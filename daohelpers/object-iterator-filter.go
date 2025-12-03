package daohelpers

// FilterIterator is a type that filters elements from an
// ObjectIterator based on a provided predicate function.
type FilterIterator[R any] struct {
	iter ObjectIterator[R]
	fn   func(r R) bool
}

// Next retrieves the next element from the iterator that satisfies the
// filter function or returns an error if iteration ends.
func (mi *FilterIterator[R]) Next() (R, error) {
	for {
		r, err := mi.iter.Next()
		if err != nil {
			var s R
			return s, err
		}

		if mi.fn(r) {
			return r, nil
		}
	}
}

// Close releases resources associated with the FilterIterator and ensures
// proper cleanup by calling the underlying iterator's Close method.
func (mi *FilterIterator[R]) Close() error {
	return mi.iter.Close()
}

// Filter creates a FilterIterator that filters elements of the provided
// ObjectIterator based on the given predicate function. The function panics
// if either the iterator or predicate is nil.
func Filter[R any](iter ObjectIterator[R], fn func(r R) bool) *FilterIterator[R] {
	if iter == nil || fn == nil {
		panic("Both parameters are required")
	}

	return &FilterIterator[R]{
		iter: iter,
		fn:   fn,
	}
}
