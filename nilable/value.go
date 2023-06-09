package nilable

// New creates a new Nilable for an
func New[T any]() Nilable[T] {
	return NewWrapper[T](&Value[T]{})
}

// Value is the base version of Nilable. It stores a value of any type
type Value[T any] struct {
	Base
	Data T
}

// Value returns the stored value without checking if it is Null. This may
// be a zero value or nil if IsNull is true.
func (p *Value[T]) Value() T {
	return p.Data
}

// SetNil sets whether the object is nil. It is self returning.
func (p *Value[T]) SetNil(b bool) Nilable[T] {
	p.Nil = b
	return p
}

// SetValue sets the value, regardless of whether the value is null. It is
// self returning.
func (p *Value[T]) SetValue(value T) Nilable[T] {
	p.Data = value
	return p
}
