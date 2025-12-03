package postgres

import (
	"database/sql/driver"

	"github.com/monstercat/golib/nilable"
)

// Nilable represents a type that can explicitly specify whether its value is nil.
type Nilable interface {
	// IsNil checks if the value is null
	IsNil() bool
}

// NilableData is a generic interface that combines the Nilable interface with
// a method to retrieve its data value. T must be a comparable type and
// represents the underlying data type managed by the implementation.
type NilableData[T comparable] interface {
	Nilable

	Data() T
}

// IsValueNil determines whether the provided Nilable value is nil by invoking
// the IsNil method on the interface.
func IsValueNil(value any) bool {
	if v, ok := value.(Nilable); ok {
		return v.IsNil()
	}
	return false
}

// TreatNilableData resolves and returns the data value from a NilableData
// object if applicablle, otherwise returns the input as is.
func TreatNilableData[T comparable](value any) any {
	if v, ok := value.(NilableData[T]); ok {
		return v.Data()
	}
	return value
}

// Nil wraps a pointer to a comparable type to automatically set as the zero
// value when nil is present.
type Nil[T comparable] struct {
	a *T
}

func NewNil[T comparable](a *T) *Nil[T] { return &Nil[T]{a} }

// Scan implements sql.Scanner
func (n *Nil[T]) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	v := nilable.New[T]()
	if err := v.Scan(value); err != nil {
		return err
	}

	*n.a = v.Data()
	return nil
}

// Value implements driver.Valuer
func (n *Nil[T]) Value() (driver.Value, error) {
	return *n.a, nil
}
