package nilable

// equals is an interface for allowing comparisons. It is used internally by
// the package.
type equals[T any] interface {
	Equal(T) bool
}
