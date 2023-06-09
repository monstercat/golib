package nilable

// New returns a new wrapper.
func New[T any]() *Wrapper[T] {
	return &Wrapper[T]{}
}

// Wrapper wraps any nullable type. If provides functionality for adding
// functionality such as adding JSON marshalling and unmarshalling logic. When
// creating a new nullable, this wrapper is the standard.
type Wrapper[T any] struct {
	// The underlying nullable.
	Nilable[T]
}

// With uses the provided Nilable as the base Nilable for the wrapper.
func (w *Wrapper[T]) With(n Nilable[T]) *Wrapper[T] {
	w.Nilable = n
	return w
}

// Unwrap returns the Nilable underneath.
func (w *Wrapper[T]) Unwrap() Nilable[T] {
	return w.Nilable
}

// Wraps defines a type of nullable which wraps another nullable. For example,
// JSON and Scanner are both wrapping Nullables. In general, a wrapping Nilable
// should have a Nilable[T] anonymous field such that it inherits the
// additional functionality provided by each wrapped layer.
type Wraps[T any] interface {
	// Unwrap returns the Nilable underneath.
	Unwrap() Nilable[T]
}

// Find returns the Nilable of the given type within a chain of Nilables.
func Find[T any, K Nilable[T]](n Nilable[T]) Nilable[T] {
	// If the current nullable is K, we should just return
	// true.
	if _, ok := n.(K); ok {
		return n
	}
	// If it cannot be unwrapped, return false. We are at the end of the chain.
	u, ok := n.(Wraps[T])
	if !ok {
		return nil
	}
	return Find[T, K](u.Unwrap())
}

// Has detects, given a Nilable, if any Nilable within the chain is a K.
func Has[T any, K Nilable[T]](n Nilable[T]) bool {
	return Find[T, K](n) != nil
}

// Add wraps the provided Nilable with the requested K if and only if the
// provided Nilable does not already include K. Note that the provided Nilable
// must be a Wrapper.
func Add[T any, K Nilable[T]](
	n Nilable[T],
	fn func(n Nilable[T]) K,
	opts ...Option[T],
) {
	w, ok := n.(*Wrapper[T])
	if !ok {
		return
	}
	if v := Find[T, K](w.Nilable); v != nil {
		ReplaceOptions(v, opts...)
		return
	}
	w.Nilable = fn(w.Nilable)
	ReplaceOptions(w.Nilable, opts...)
}

// HasJSON returns true if one of the Nullables in the chain of Nilable is
// a *JSON type.
func HasJSON[T any](n Nilable[T]) bool {
	return Has[T, *JSON[T]](n)
}

// AddJSON ensures that the provided Nilable includes a JSON. This only works
// if the provided Nilable *is* a Wrapper. Otherwise, it is ignored.
func AddJSON[T any](n Nilable[T], opts ...Option[T]) {
	Add[T, *JSON[T]](n, NewJSON[T], opts...)
}

// HasScanner returns true if one of the Nullables in the chain of Nilable is
// a *Scanner type.
func HasScanner[T any](n Nilable[T]) bool {
	return Has[T, *Scanner[T]](n)
}

// AddScanner ensures that the provided Nilable includes a JSON. This only
// works if the provided Nilable *is* a Wrapper. Otherwise, it is ignored.
func AddScanner[T any](n Nilable[T], opts ...Option[T]) {
	Add[T, *Scanner[T]](n, NewScanner[T], opts...)
}
