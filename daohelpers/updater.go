package daohelpers

import (
	"golang.org/x/exp/constraints"
)

// Updating provides a quick way to define functions that create or update
type Updating[ID string | constraints.Integer] interface {
	// Insert creates the provided data, and returns the ID
	// If insert operation fails due to duplicated data, return ErrDuplicate
	Insert() (ID, error)

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

// Overriding is a generic interface that allows setting override values via a
// map of key and value pairs.
type Overriding[T any] interface {
	// SetOverrideRules sets the override rules for the Updater.Set method calls.
	//
	// **Usage Example**
	//
	//	var payload partial.Payload[model.AllocationSet]
	//	c.ShouldBindJSON(&payload)
	//
	//	omitted := payload.OmittedFields(partial.GetTagValues("db"))
	//
	//	err := bl.Providers().AllocationSets().
	//		Use(tx).
	//		WithId(payload.Id).
	//		// Setting override rules for partial update in the db.
	//		SetOverrideRules(dao.OmitFields(omitted...)).
	//		// ... implement all allowed setter method chaining calls as is.
	//		Update()
	//
	// **Implementation Example**
	//
	//	func (p *AllocationSetPostgres) SetOverrideRules(rules map[string]any) AllocationWriter {
	//		// custom logic to modify, filter, or transform the fields and rules for
	//		// the allocation DAO
	//		p.SetOverrides(rules)
	//		return p
	//	}
	SetOverrideRules(overrides map[string]any) T
}
