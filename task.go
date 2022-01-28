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

// Task represents a unit of work.
type Task interface {
	// Execute executes the task. This is called by the scheduler to start the task.
	Execute()
	// Wait will return immediately if the task is complete. It will block if the task is running.
	Wait() error
	// Result returns the result. It will not block and will return immediately.
	Result() interface{}
	// Error returns the error It will not block and will return immediately.
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
	mutex       sync.RWMutex // currently this is a shared mutex for all state, switch to individual?
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
	// cache execution
	if t.IsCompleted() {
		return
	}

	// cleanup the channel after execution completes
	defer close(t.doneCh)

	// execute the delegate
	result, err := t.errFuncWith(t.state)

	// notify subscribers
	if err != nil {
		t.setError(err)
		t.setStatus(StatusFaulted)
		t.NotifyError(err)
	} else {
		t.setResult(result)
		t.setStatus(StatusSuccess)
		t.NotifyNext(result)
	}
	t.doneCh <- struct{}{}
}

func (t *task) Wait() error {
	// cache execution
	if t.IsCompleted() {
		return t.Error()
	}

	// this allows for the task to be canceled. The main work is done in execute.
	select {
	case <-t.doneCh:
		return t.Error()
	case <-t.context.Done():
		err := t.context.Err()
		t.setError(err)
		t.setStatus(StatusCanceled)
		t.NotifyCanceled(err)
		return err
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
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.result
}

func (t *task) setResult(result interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.result = result
}

func (t *task) Error() error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.err
}

func (t *task) setError(err error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.err = err
}

func (t *task) IsCompleted() bool {
	switch t.Status() {
	case StatusCanceled, StatusFaulted, StatusSuccess:
		return true
	default:
		return false
	}
}

func (t *task) IsCanceled() bool {
	return t.Status() == StatusCanceled
}

func (t *task) IsFaulted() bool {
	return t.Status() == StatusFaulted
}

func (t *task) Status() TaskStatus {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.status
}

func (t *task) setStatus(status TaskStatus) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.status = status
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

// Delay calls time.Sleep with the given duration
func Delay(duration time.Duration, options ...RunOption) ObservableTask {
	return RunAction(func() {
		time.Sleep(duration)
	}, options...)
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
