package task

// RunAction runs the given action function with the supplied RunOptions
// An Action is a function with no arguments and no returns
func RunAction(action Action, options ...RunOption) Task {
	errFuncWith := func(interface{}) (interface{}, error) {
		action()
		return nil, nil
	}
	return RunErrFuncWith(errFuncWith, options...)
}

// RunActionWith runs the given action function with the supplied RunOptions
// An ActionWith is a function with one interface argument and no returns. The interface
// argument can be supplied with task.WithState(state) in the options parameter list.
func RunActionWith(actionWith ActionWith, options ...RunOption) Task {
	errFuncWith := func(state interface{}) (interface{}, error) {
		actionWith(state)
		return nil, nil
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunErrAction(errAction ErrAction, options ...RunOption) Task {
	errFuncWith := func(interface{}) (interface{}, error) {
		return nil, errAction()
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunErrActionWith(errActionWith ErrActionWith, options ...RunOption) Task {
	errFuncWith := func(state interface{}) (interface{}, error) {
		return nil, errActionWith(state)
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunFunc(f Func, options ...RunOption) Task {
	errFuncWith := func(interface{}) (interface{}, error) {
		return f(), nil
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunFuncWith(funcWith FuncWith, options ...RunOption) Task {
	errFuncWith := func(state interface{}) (interface{}, error) {
		return funcWith(state), nil
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunErrFunc(f ErrFunc, options ...RunOption) Task {
	errFuncWith := func(interface{}) (interface{}, error) {
		return f()
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunErrFuncWith(errFuncWith ErrFuncWith, options ...RunOption) Task {
	// create the task
	t := new(errFuncWith)

	// apply operations
	for _, opt := range options {
		opt(t)
	}
	t.scheduler.Queue(t)
	return t
}
