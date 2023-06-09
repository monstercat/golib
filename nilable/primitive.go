package nilable

import "golang.org/x/exp/constraints"

// NewPrimitive creates a new Nilable for an object that is a primitive,
// assuming that the object is not null.
func NewPrimitive[T primitive]() Nilable[T] {
	return &Primitive[T]{}
}

// primitive is a union of all types which can be considered primitive. This is
// to take advantage of GOLANG compilation whereby generics which are primitives
// are compiled as separate structs rather than a pointer to a pointer.
type primitive interface {
	constraints.Ordered
	constraints.Complex
	~[]byte
}

// Primitive stores a primitive data type along with a flag that defines it as
// a nullable type.
type Primitive[T primitive] struct {
	Base
	Data T
}

// Value returns the stored value without checking if it is Null. This may
// be a zero value or nil if IsNull is true.
func (p *Primitive[T]) Value() T {
	return p.Data
}

// SetNil sets whether the object is nil. It is self returning.
func (p *Primitive[T]) SetNil(b bool) Nilable[T] {
	p.Nil = b
	return p
}

// SetValue sets the value, regardless of whether the value is null. It is
// self returning.
func (p *Primitive[T]) SetValue(value T) Nilable[T] {
	p.Data = value
	return p
}
