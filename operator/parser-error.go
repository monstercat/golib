package operator

import "fmt"

// ParseMapKeyError is an error related to parsing of a map key. It wraps the error returned from
// ParserRegexpDelegate.ParseMapKey
type ParseMapKeyError struct {
	// The error in question
	Base error

	// InputKey which caused the error.
	InputKey string
}

// Error string
func (e *ParseMapKeyError) Error() string {
	return fmt.Sprintf("Input key '%s' has been ignored. %s", e.InputKey, e.Base)
}

// ParseRegexpError occurs when regexp could not be parsed properly for Parser.Parse.
type ParseRegexpError struct {
	// The error in question
	Base error
}

// Error string
func (e *ParseRegexpError) Error() string {
	return fmt.Sprintf("Regexp could not be compiled. %s", e.Base)
}

// ParseMatchError is an error related to the parsing of a match. It wraps the error returned from
// ParserRegexpDelegate.ParseMatch
type ParseMatchError struct {
	// Matches that were returned, if any.
	Matches []string

	// The error in question
	Base error
}

// Error string
func (e *ParseMatchError) Error() string {
	return fmt.Sprintf("Could not parse provided match. %s", e.Base)
}