package task

// NewAction creates a new unstarted task with the given delegate and run options
func NewAction(action Action, options ...RunOption) Task {
	errFuncWith := func(interface{}) (interface{}, error) {
		action()
		return nil, nil
	}
	return NewErrFuncWith(errFuncWith, options...)
}

// NewActionWith creates a new unstarted task with the given delegate and run options
func NewActionWith(actionWith ActionWith, options ...RunOption) Task {
	errFuncWith := func(state interface{}) (interface{}, error) {
		actionWith(state)
		return nil, nil
	}
	return NewErrFuncWith(errFuncWith, options...)
}

// NewErrAction creates a new unstarted task with the given delegate and run options
func NewErrAction(errAction ErrAction, options ...RunOption) Task {
	errFuncWith := func(interface{}) (interface{}, error) {
		return nil, errAction()
	}
	return NewErrFuncWith(errFuncWith, options...)
}

// NewErrActionWith creates a new unstarted task with the given delegate and run options
func NewErrActionWith(errActionWith ErrActionWith, options ...RunOption) Task {
	errFuncWith := func(state interface{}) (interface{}, error) {
		return nil, errActionWith(state)
	}
	return NewErrFuncWith(errFuncWith, options...)
}

// NewFunc creates a new unstarted task with the given delegate and run options
func NewFunc(f Func, options ...RunOption) Task {
	errFuncWith := func(interface{}) (interface{}, error) {
		return f(), nil
	}
	return NewErrFuncWith(errFuncWith, options...)
}

// NewFuncWith creates a new unstarted task with the given delegate and run options
func NewFuncWith(funcWith FuncWith, options ...RunOption) Task {
	errFuncWith := func(state interface{}) (interface{}, error) {
		return funcWith(state), nil
	}
	return NewErrFuncWith(errFuncWith, options...)
}

// NewErrFunc creates a new unstarted task with the given delegate and run options
func NewErrFunc(f ErrFunc, options ...RunOption) Task {
	errFuncWith := func(interface{}) (interface{}, error) {
		return f()
	}
	return NewErrFuncWith(errFuncWith, options...)
}

// NewErrFuncWith creates a new unstarted task with the given delegate and run options
func NewErrFuncWith(errFuncWith ErrFuncWith, options ...RunOption) Task {
	// create the task
	t := new(errFuncWith)

	// apply operations
	for _, opt := range options {
		opt(t)
	}

	// return unstarted task
	return t
}
