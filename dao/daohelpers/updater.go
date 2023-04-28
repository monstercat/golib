package daohelpers

// Updating provides a quick way to define functions that create or update
type Updating interface {
	// Insert creates the provided data, and returns the ID
	Insert() (string, error)

	// Update updates the data. If no conditions are applied lm.ErrNoConditions should
	// be returned.
	Update() error
}

// Deleting provides a quick way to define Delete functionality.
type Deleting interface {
	// Delete deletes the data. If no conditions are applied, lm.ErrNoConditions
	// should be returned.
	Delete() error
}
