package postgres

// Overrider implmeents daohelpers.Overriding
type Overrider[T any, ID any] struct {
	// Updater is what is actually performing the update. Internally, it has
	// override logic in place. This override logic is what will be used to
	// modify the data before it is sent to the database. However, the data
	// in the updater is purely a postgres implementation, while the
	// daohelpers.Overiding is an interface agnostic to the type of database.
	//
	// Thus, this struct serves as a proxy to convert the fields so they
	// match the fields being set.
	Updater *Updater[ID]

	// FieldMapper maps the fields in the payload to the fields in the database.
	FieldMapper map[string]string

	// ReturnVal is the value to return from the SetOverrideRules method. This
	// is usually the *Writer interface used for chaining.
	ReturnVal T
}

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
func (o *Overrider[T, ID]) SetOverrideRules(overrides map[string]any) T {
	if o.FieldMapper == nil {
		o.Updater.SetOverrides(overrides)
		return o.ReturnVal
	}

	mapped := make(map[string]any, len(overrides))
	for k, v := range overrides {
		x, ok := o.FieldMapper[k]
		if ok {
			mapped[x] = v
		}
	}

	o.Updater.SetOverrides(mapped)
	return o.ReturnVal
}
