package errors

import "strings"

type Errors []error

func (e Errors) Error() string {
	strs := make([]string, 0, len(e))
	for _, ee := range e {
		strs = append(strs, ee.Error())
	}
	return strings.Join(strs, "\r\n")
}

func (e *Errors) AddError(err error) {
	*e = append(*e, err)
}
