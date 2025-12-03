package postgres

// Nullable processes a nullable value for an Updater, setting nil if
// the value is nil or transforming it otherwise.
//
// Usage Example:
// The following example sets the label_scope_id column to nil if the value is
// nil otherwises sets the value with id.Data()
//
//	func (p *SheetPostgres) SetLabelScopeId(id nilable.Value[string]) SheetWriter {
//		p.Set(TableSheetColumnLabelScopeId, id, Nullable[string])
//		return p
//	}
func Nullable[T comparable](column string, value any) (any, bool) {
	if IsValueNil(value) {
		// if the value is null, update the column with nil and interrupt from
		// further update procedures on the column.
		return nil, true
	}
	return TreatNilableData[T](value), false
}

// OmitIfNil returns an omitted placeholder and interrupt signal, if the value
// is nil; otherwise, it returns the value and false.
//
// Usage Example:
// The following example omits the posted_date column update if the value is nil,
// since the field is not clearable once set.
//
//	func (p *SheetPostgres) SetPostedDate(dt nilable.Value[time.Time]) SheetWriter {
//		p.Set(TableSheetColumnPostedDate, dt, pg.OmitIfNil)
//		return p
//	}
func OmitIfNil[T comparable](column string, value any) (any, bool) {
	if IsValueNil(value) {
		return OmmitedValue{}, false
	}
	return TreatNilableData[T](value), false
}
