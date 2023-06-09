package nilable

// NewScanner wraps a nullable with the Scanner version.
func NewScanner[T any](n Nilable[T]) *Scanner[T] {
	// In the case it is already Scanner wrapped, just ignore and
	// return original.
	if v, ok := n.(*Scanner[T]); ok {
		return v
	}
	return &Scanner[T]{
		Nilable: n,
	}
}

// Scanner implements the scanner interface.
type Scanner[T any] struct {
	Nilable[T]
	Optionable[T]
}

// Unwrap returns the Nilable underneath.
func (s *Scanner[T]) Unwrap() Nilable[T] {
	return s.Nilable
}
