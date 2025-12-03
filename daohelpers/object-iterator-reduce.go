package daohelpers

import "golang.org/x/exp/constraints"

// ReduceIterator is a generic type used to aggregate results from an ObjectIterator[R] using a custom aggregation function.
// The `iter` field stores the underlying ObjectIterator[R] to fetch items.
// The `fn` field defines the logic for aggregating results and when to emit a value.
type ReduceIterator[R any, T any] struct {
	iter ObjectIterator[R]
	fn   func(agg T, item R) (T, bool)
}

// Next retrieves the next aggregated result from the iterator or
// returns an error if no more items are available.
func (r *ReduceIterator[R, T]) Next() (T, error) {
	var t T
	for {
		orig, err := r.iter.Next()
		if err != nil {
			return t, err
		}
		var ok bool
		t, ok = r.fn(t, orig)
		if ok {
			var tt T
			t = tt
			return t, nil
		}
	}
}

// Close terminates the underlying iterator and releases associated resources.
// Returns an error if the closure fails.
func (r *ReduceIterator[R, T]) Close() error {
	return r.iter.Close()
}

// Reduce creates a ReduceIterator to aggregate results from an ObjectIterator using a custom aggregation function.
// Panics if either the ObjectIterator or aggregation function is nil.
//
// The aggregate function should return the new version of the aggregate, as well as a boolean indicate whether
// we are ready to return a value. If no value is ever returned, the last value will be returned automatically
// with iterator.Done as the error.
func Reduce[R any, T any](iter ObjectIterator[R], fn func(agg T, item R) (T, bool)) *ReduceIterator[R, T] {
	if iter == nil || fn == nil {
		panic("Both parameters are required")
	}

	return &ReduceIterator[R, T]{
		iter: iter,
		fn:   fn,
	}
}

type numeric interface {
	constraints.Integer | constraints.Float
}

// Min computes the minimum value in a numeric ObjectIterator and returns it along with any encountered error.
// Returns an error if the iterator fails or contains no elements.
// Panics if the provided ObjectIterator is nil.
// Requires elements in the iterator to implement the numeric interface (constraints.Integer or constraints.Float).
func Min[R numeric](iter ObjectIterator[R]) (R, error) {
	var firstValueMet bool
	r := Reduce[R, R](iter, func(agg R, item R) (R, bool) {
		if !firstValueMet {
			firstValueMet = true
			return item, false
		}
		if item < agg {
			return item, false
		}
		return agg, false
	})

	// The above should always only return 1 number.
	return First[R](r)
}
