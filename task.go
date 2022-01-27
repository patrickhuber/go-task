package task

import (
	"context"
	"io"
	"sync"
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
	tracker     Tracker
	mutex       sync.RWMutex
}

func new(errFuncWith ErrFuncWith) *task {
	return &task{
		context:     context.TODO(),
		status:      StatusCreated,
		scheduler:   DefaultScheduler(),
		errFuncWith: errFuncWith,
		tracker:     NewTracker(),
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
		t.setStatus(StatusFaulted)
		t.NotifyError(err)
	} else {
		t.result = result
		t.setStatus(StatusSuccess)
		t.NotifyNext(result)
	}
	t.doneCh <- struct{}{}
}

func (t *task) Wait() error {
	if t.IsCompleted() {
		return t.err
	}

	select {
	case <-t.doneCh:
		return t.err
	case <-t.context.Done():
		t.err = t.context.Err()
		t.setStatus(StatusCanceled)
		t.NotifyCanceled(t.err)
		return t.err
	}
}

func (t *task) NotifyNext(value interface{}) {
	t.tracker.NotifyNext(value)
	t.tracker.NotifyCompleted()
}

func (t *task) NotifyError(err error) {
	t.tracker.NotifyError(err)
	t.tracker.NotifyCompleted()
}

func (t *task) NotifyCanceled(err error) {
	t.tracker.NotifyCanceled(err)
	t.tracker.NotifyCompleted()
}

func (t *task) Result() interface{} {
	return t.result
}

func (t *task) Error() error {
	return t.err
}

func (t *task) IsCompleted() bool {
	switch t.Status() {
	case StatusCanceled, StatusFaulted, StatusSuccess:
		return true
	default:
		return false
	}
}

func (t *task) setStatus(status TaskStatus) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.status = status
}

func (t *task) IsCanceled() bool {
	return t.Status() == StatusCanceled
}

func (t *task) IsFaulted() bool {
	return t.Status() == StatusFaulted
}

func (t *task) Status() TaskStatus {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.status
}

func (t *task) Subscribe(o Observer) io.Closer {
	return t.tracker.Subscribe(o)
}

func (t *task) Unsubscribe(o Observer) {
	t.tracker.Unsubscribe(o)
}

// Completed returns a completed task in the StatusSuccess state
func Completed() ObservableTask {
	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}
	return &task{
		status: StatusSuccess,
		doneCh: doneCh,
	}
}

// FromResult returns a completed task in the StatusSuccess state with the given result
func FromResult(result interface{}) ObservableTask {
	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}
	return &task{
		result: result,
		status: StatusSuccess,
		doneCh: doneCh,
	}
}

// FromError returns a completed task in StatusFaulted state with the given error
func FromError(err error) ObservableTask {
	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}
	return &task{
		err:    err,
		status: StatusFaulted,
		doneCh: doneCh,
	}
}

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
