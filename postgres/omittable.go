package postgres

// Omittable is an interface for types that can signify if their value should
// be omitted.
type Omittable interface {
	IsOmitted() bool
}

// IsValueOmitted checks if the given value implements the Omittable interface
// and returns true if it should be omitted.
func IsValueOmitted(o any) bool {
	if v, ok := o.(Omittable); ok {
		return v.IsOmitted()
	}
	return false
}

// OmmitedValue represents a placeholder type used to indicate that a value has
// been explicitly omitted.
type OmmitedValue struct{}

// IsOmitted checks whether the value is explicitly marked as omitted, always
// returning true.
func (OmmitedValue) IsOmitted() bool {
	return true
}
