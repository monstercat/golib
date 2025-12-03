package daohelpers

// MapIterator maps values of type R to type S
type MapIterator[R any, S any] struct {
	iter ObjectIterator[R]
	fn   func(r R) S
}

func (mi *MapIterator[R, S]) Next() (S, error) {
	r, err := mi.iter.Next()
	if err != nil {
		var s S
		return s, err
	}

	return mi.fn(r), nil
}

func (mi *MapIterator[R, S]) Close() error {
	return mi.iter.Close()
}

// Map maps one iterator to another using a function.
func Map[R any, S any](iter ObjectIterator[R], fn func(r R) S) ObjectIterator[S] {
	if iter == nil || fn == nil {
		panic("Both parameters are required")
	}

	return &MapIterator[R, S]{
		iter: iter,
		fn:   fn,
	}
}
