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
}

func AppendError(errors ...error) AggregateError {
	if len(errors) == 0 {
		return nil
	}
	if errors[0] == nil {
		return &aggregateError{
			errors: errors[1:],
		}
	}
	if a, ok := errors[0].(AggregateError); ok {
		a.Append(errors[1:]...)
		return a
	}
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

func (err *aggregateError) ErrOrNil() error {
	if len(err.errors) == 0 {
		return nil
	}
	return err
}
