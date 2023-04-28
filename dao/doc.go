// Package dao provides utility structs, interfaces and methods related to
// Data Access Objects (DAOs). DAOs are a paradigm for data layer abstraction
// which uses self-returning functions to build parameters, ending in a CRUD
// function.
//
// Typically, parameter functions start with either `With` or `Set`. A `With`
// function is used to restrict the operation to a certain subset of the
// data, while a `Set` function is typically used to set data for an update
// or a create method.
//
// The following sub-packages are provided to ease a developer's burden when
// creating DAO interfaces and implementations. Please read the documentation
// specific to each sub-package for more details.
// - daohelpers
// - postgres
// - transaction
// - testdao
package dao
