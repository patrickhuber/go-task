package task

import (
	"context"
	"io"
	"time"
)

type TaskStatus string

const (
	StatusCreated  TaskStatus = "created"
	StatusSuccess  TaskStatus = "success"
	StatusFaulted  TaskStatus = "faulted"
	StatusCanceled TaskStatus = "canceled"
)

type Task interface {
	Execute()
	Wait() error
	Result() interface{}
	Error() error
	IsCompleted() bool
	IsFaulted() bool
	IsCanceled() bool
	Status() TaskStatus
}

type ObservableTask interface {
	Observable
	Task
}

type task struct {
	status      TaskStatus
	result      interface{}
	err         error
	doneCh      chan struct{}
	context     context.Context
	scheduler   Scheduler
	errFuncWith ErrFuncWith
	state       interface{}
	observers   []Observer
}

func new(errFuncWith ErrFuncWith) *task {
	return &task{
		context:     context.TODO(),
		status:      StatusCreated,
		scheduler:   DefaultScheduler(),
		errFuncWith: errFuncWith,
		// make this buffered to avoid blocking the calling routine
		doneCh: make(chan struct{}, 1),
	}
}

type RunOption func(t *task)

func WithContext(ctx context.Context) RunOption {
	return func(t *task) {
		t.context = ctx
	}
}

func WithTimeout(timeout time.Duration) RunOption {
	return func(t *task) {
		t.context, _ = context.WithTimeout(t.context, timeout)
	}
}

func WithScheduler(s Scheduler) RunOption {
	return func(t *task) {
		t.scheduler = s
	}
}

func WithState(state interface{}) RunOption {
	return func(t *task) {
		t.state = state
	}
}

func (t *task) Execute() {
	if t.IsCompleted() {
		return
	}
	result, err := t.errFuncWith(t.state)
	if err != nil {
		t.err = err
		t.status = StatusFaulted
		t.NotifyError(err)
	} else {
		t.result = result
		t.status = StatusSuccess
		t.NotifyNext(result)
	}
	t.doneCh <- struct{}{}
}

func (t *task) Wait() error {
	// result is cached? return it
	if t.IsCompleted() {
		return t.err
	}

	select {
	case <-t.doneCh:
		return t.err
	case <-t.context.Done():
		t.err = t.context.Err()
		t.status = StatusCanceled
		t.NotifyCanceled(t.err)
		return t.err
	}
}

func (t *task) NotifyNext(value interface{}) {
	for _, o := range t.observers {
		o.OnNext(value)
		// tasks are completed when they receive any message
		o.OnCompleted()
	}
}

func (t *task) NotifyError(err error) {
	for _, o := range t.observers {
		o.OnError(err)
		// tasks are completed when they receive any message
		o.OnCompleted()
	}
}

func (t *task) NotifyCanceled(err error) {
	for _, o := range t.observers {
		o.OnCanceled(err)
		// tasks are completed when they receive any message
		o.OnCompleted()
	}
}

func (t *task) Result() interface{} {
	return t.result
}

func (t *task) Error() error {
	return t.err
}

func (t *task) IsCompleted() bool {
	switch t.status {
	case StatusCanceled, StatusFaulted, StatusSuccess:
		return true
	default:
		return false
	}
}

func (t *task) IsCanceled() bool {
	return t.status == StatusCanceled
}

func (t *task) IsFaulted() bool {
	return t.status == StatusFaulted
}

func (t *task) Status() TaskStatus {
	return t.status
}

func (t *task) Subscribe(o Observer) io.Closer {
	contains := false
	for _, observer := range t.observers {
		if o == observer {
			contains = true
			break
		}
	}
	if !contains {
		t.observers = append(t.observers, o)
	}
	return newSubscription(o, t.observers)
}

func Completed() ObservableTask {
	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}
	return &task{
		status: StatusSuccess,
		doneCh: doneCh,
	}
}

func FromResult(result interface{}) ObservableTask {
	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}
	return &task{
		result: result,
		status: StatusSuccess,
		doneCh: doneCh,
	}
}

func FromError(err error) ObservableTask {
	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}
	return &task{
		err:    err,
		status: StatusFaulted,
		doneCh: doneCh,
	}
}

func RunAction(action Action, options ...RunOption) ObservableTask {
	errFuncWith := func(interface{}) (interface{}, error) {
		action()
		return nil, nil
	}
	return RunErrFuncWith(errFuncWith, options...)
}

func RunActionWith(actionWith ActionWith, options ...RunOption) ObservableTask {
	errFuncWith := func(state interface{}) (interface{}, error) {
		actionWith(state)
		return nil, nil
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
