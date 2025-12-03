package daohelpers

// DataSetter is an interface to set data based on a key/field name.
type DataSetter interface {
	Set(field string, data any, opts ...UpdateRule)
}

// MapDataSetter implements DataSetter using a map.
type MapDataSetter struct {
	M map[string]any
}

func NewMapDataSetter() *MapDataSetter {
	return &MapDataSetter{
		M: make(map[string]any),
	}
}

// Set sets the data in the field. It is required for implementing
// DataSetter.
func (s *MapDataSetter) Set(field string, data any, opts ...UpdateRule) {
	s.M[field] = data
}

// UpdateRule defines rule for data update operations performed by data access
// writer object.
type UpdateRule func(column string, value any) (treated any, interrupt bool)
