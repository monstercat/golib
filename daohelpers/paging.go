package daohelpers

// Paging interface is useful for denoting functions related to paging in DAOs.
type Paging[R any] interface {
	// WithLimit should introduce a limit to the DAO results or limit the amount of updated items
	WithLimit(limit uint64) R

	// GetLimit returns the set limit
	GetLimit() uint64

	// WithOffset should introduce an offset to the DAO results or offset the items which are updated
	WithOffset(offset uint64) R

	// GetOffset returns the set offset.
	GetOffset() uint64
}

// Sorting interface is useful for denoting functions related to sorting in DAOs.
type Sorting[R any, T any] interface {
	// WithSort sorts the items by certain pre-specified fields.
	WithSort(xs ...T) R
}
