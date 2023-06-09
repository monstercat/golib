package nilable

// Base implements half of Nilable.
type Base struct {
	Nil bool
}

// IsNil returns true if the value is null
func (b *Base) IsNil() bool {
	return b.Nil
}
