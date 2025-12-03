package daohelpers

// BatchLoaderEntry is an object returned by the BatchLoader to allow addition
// of items into the Batch. It should include the Commit function, which, when
// called, actually performs the addition of data into the BatchLoader. Any
// other data, which can be passed through any other functions.
type BatchLoaderEntry interface {
	// Commit pushes the new entry into the batch loader.
	Commit() error
}

// BatchLoader starts a loader which can load data in a batch.
type BatchLoader[T BatchLoaderEntry] interface {
	// New creates a new BatchLoaderEntry.
	New() T

	// Flush should commit any fully loaded data to the backend.
	Flush() error

	// Len should return the number of entries not yet flushed.
	Len() int
}
