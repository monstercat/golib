package nilable

// Nilable is a generic interface which all structs in this package implement.
type Nilable[T any] interface {
	// IsNil returns true if the value is null
	IsNil() bool

	// Value returns the stored value without checking if it is Null. This may
	// be a zero value or nil if IsNull is true.
	Value() T

	// SetNil sets whether the object is nil. It is self returning.
	SetNil(b bool) Nilable[T]

	// SetValue sets the value, regardless of whether the value is null. It is
	// self returning.
	SetValue(value T) Nilable[T]
}
