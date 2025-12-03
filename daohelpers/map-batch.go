package daohelpers

// MapBatchLoader is a BatchLoader with an Add function that takes in a
// map[string]any
type MapBatchLoader[T BatchLoaderEntry] interface {
	BatchLoader[T]

	// Add adds data in map form.
	Add(data map[string]any) error
}
