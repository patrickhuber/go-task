package task

type Action func()
type ActionWith func(interface{})
type ErrAction func() error
type ErrActionWith func(interface{}) error
type Func func() interface{}
type FuncWith func(interface{}) interface{}
type ErrFunc func() (interface{}, error)
type ErrFuncWith func(interface{}) (interface{}, error)

// RunAction runs the given action function with the supplied RunOptions
// An Action is a function with no arguments and no returns
func RunAction(action Action, options ...RunOption) ObservableTask {
	errFuncWith := func(interface{}) (interface{}, error) {
		action()
		return nil, nil
	}
	return RunErrFuncWith(errFuncWith, options...)
}

// RunActionWith runs the given action function with the supplied RunOptions
// An ActionWith is a function with one interface argument and no returns. The interface
// argument can be supplied with task.WithState(state) in the options parameter list.
func RunActionWith(actionWith ActionWith, options ...RunOption) ObservableTask {
	errFuncWith := func(state interface{}) (interface{}, error) {
		actionWith(state)
		return nil, nil
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunErrAction(errAction ErrAction, options ...RunOption) ObservableTask {
	errFuncWith := func(interface{}) (interface{}, error) {
		return nil, errAction()
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunErrActionWith(errActionWith ErrActionWith, options ...RunOption) ObservableTask {
	errFuncWith := func(state interface{}) (interface{}, error) {
		return nil, errActionWith(state)
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunFunc(f Func, options ...RunOption) ObservableTask {
	errFuncWith := func(interface{}) (interface{}, error) {
		return f(), nil
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunFuncWith(funcWith FuncWith, options ...RunOption) ObservableTask {
	errFuncWith := func(state interface{}) (interface{}, error) {
		return funcWith(state), nil
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunErrFunc(f ErrFunc, options ...RunOption) ObservableTask {
	errFuncWith := func(interface{}) (interface{}, error) {
		return f()
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunErrFuncWith(errFuncWith ErrFuncWith, options ...RunOption) ObservableTask {
	// create the task
	t := new(errFuncWith)

	// apply operations
	for _, opt := range options {
		opt(t)
	}
	t.scheduler.Queue(t)
	return t
}
