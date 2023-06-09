package nilable

import "time"

// NewTime creates a new time nullable.
func NewTime() Nilable[time.Time] {
	t := &Time{}
	return t
}

// Time stores a time value which could be nil.
type Time struct {
	Base
	Data time.Time
}

// Value returns the stored value without checking if it is Null. This may
// be a zero value or nil if IsNull is true.
func (p *Time) Value() time.Time {
	return p.Data
}

// SetNil sets whether the object is nil. It is self returning.
func (p *Time) SetNil(b bool) Nilable[time.Time] {
	p.Nil = b
	return p
}

// SetValue sets the value, regardless of whether the value is null. It is
// self returning.
func (p *Time) SetValue(value time.Time) Nilable[time.Time] {
	p.Data = value
	return p
}
