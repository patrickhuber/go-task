package task

import (
	"fmt"
)

type aggregateError struct {
	errors []error
}

type AggregateError interface {
	error
	Append(erors ...error)
	Errors() []error
}

func AppendError(errors ...error) AggregateError {
	filtered := []error{}
	// filter out nil
	for _, e := range errors {
		if e == nil {
			continue
		}
		filtered = append(filtered, e)
	}
	errors = filtered

	if len(errors) == 0 {
		return nil
	}

	// is the first error an aggregate error?
	if a, ok := errors[0].(AggregateError); ok {
		if len(errors) > 1 {
			a.Append(errors[1:]...)
		}
		return a
	}

	// the first error is not an aggregate error, so create a new aggregate error and append all
	a := &aggregateError{
		errors: []error{},
	}
	a.Append(errors...)
	return a
}

func (err *aggregateError) Append(errors ...error) {
	err.errors = append(err.errors, errors...)
}

func (err *aggregateError) Error() string {
	outStr := ""
	for _, e := range err.errors {
		outStr = fmt.Sprintln(e.Error())
	}
	return outStr
}

func (err *aggregateError) Errors() []error {
	return err.errors
}
