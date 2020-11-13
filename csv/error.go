package csv

import "fmt"

const (
	BadDateErrorFormat          = "Bad field '%s' format ('%s') was provided. For example, try 2006-01-02 15:04:05 -0700, line %d"
	BadDecimalErrorFormat       = "Bad field '%s' format ('%s') was provided. Must be a number. Line %d"
	EmptyFieldErrorFormat       = "'%s' must not be empty, line %d"
	UnreasonableTimeErrorFormat = "Date '%s' is invalid, field %s, line %d"
)

func HandleGroupedError(errs *MultipleError, err error) bool {
	if err == nil {
		return false
	}
	errs.AddError(err)
	return true
}

type HeaderError string

func (e HeaderError) Error() string {
	return string(e)
}

type ColumnMismatchError struct {
	Line     int
	Expected int
	Got      int
}

func (e ColumnMismatchError) Error() string {
	return fmt.Sprintf("Line %d column count mismatch, got %d need at least %d", e.Line, e.Got, e.Expected)
}

type Error interface {
	Error() string
	SetField(string)
	SetLine(int)
}
type BaseError struct {
	Field string
	Line  int
}

func (e *BaseError) Error() string {
	return ""
}
func (e *BaseError) SetField(str string) {
	e.Field = str
}
func (e *BaseError) SetLine(n int) {
	e.Line = n
}

func TransformCsvError(err error, field string, line int) error {
	if err == nil {
		return err
	}
	verr, ok := err.(Error)
	if !ok {
		return err
	}
	verr.SetLine(line)
	verr.SetField(field)
	return verr
}

type BadDateFormatError struct {
	BaseError
	Date string
}

func (e *BadDateFormatError) Error() string {
	return fmt.Sprintf(BadDateErrorFormat, e.Field, e.Date, e.Line)
}

type UnreasonableTimeCsvError struct {
	BaseError
	Date string
}

func (e *UnreasonableTimeCsvError) Error() string {
	return fmt.Sprintf(UnreasonableTimeErrorFormat, e.Date, e.Field, e.Line)
}

type EmptyFieldError BaseError

func (e *EmptyFieldError) Error() string {
	return fmt.Sprintf(EmptyFieldErrorFormat, e.Field, e.Line)
}

type BadDecimalFormatError struct {
	BaseError
	Value string
}

func (e *BadDecimalFormatError) Error() string {
	return fmt.Sprintf(BadDecimalErrorFormat, e.Field, e.Value, e.Line)
}

type MultipleError struct {
	Errors []error
}

func (e *MultipleError) Error() string {
	return "Multiple CSV Errors"
}

func (e *MultipleError) Return() error {
	if len(e.Errors) == 0 {
		return nil
	}
	return e
}

func (e *MultipleError) AddError(err error) {
	e.Errors = append(e.Errors, err)
}
