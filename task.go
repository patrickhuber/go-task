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

	Continuation
}

type ContinueAction func(Task)
type ContinueActionWith func(Task, interface{})
type ContinueErrAction func(Task) error
type ContinueErrActionWith func(Task, interface{}) error
type ContinueFunc func(Task) interface{}
type ContinueFuncWith func(Task, interface{}) interface{}
type ContinueErrFunc func(Task) (interface{}, error)
type ContinueErrFuncWith func(Task, interface{}) (interface{}, error)

type Continuation interface {
	ContinueAction(ContinueAction) ObservableTask
	ContinueActionWith(ContinueActionWith) ObservableTask
	ContinueErrActionWith(ContinueErrActionWith) ObservableTask
	ContinueErrAction(ContinueErrAction) ObservableTask
	ContinueFunc(ContinueFunc) ObservableTask
	ContinueFuncWith(ContinueFuncWith) ObservableTask
	ContinueErrFunc(ContinueErrFunc) ObservableTask
	ContinueErrFuncWith(ContinueErrFuncWith) ObservableTask
}

type ObservableTask interface {
	Observable
	Task
}

type task struct {
	executeOnce sync.Once
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
	t.executeOnce.Do(t.execute)
}

func (t *task) execute() {
	// if this task is already complete, return
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
		t.notifyError(err)
	} else {
		t.setResult(result)
		t.setStatus(StatusSuccess)
		t.notifyNext(result)
	}
	t.doneCh <- struct{}{}
}

func (t *task) Wait() error {
	// after completion, return the error code
	if t.IsCompleted() {
		return t.Error()
	}

	// this allows for the task to be canceled. The main work is done in execute.
	// when the channel is closed any callers blocked will read nil from t.doneCh
	select {
	case <-t.doneCh:
		return t.Error()
	case <-t.context.Done():
		err := t.context.Err()
		t.setError(err)
		t.setStatus(StatusCanceled)
		t.notifyCanceled(err)
		return err
	}
}

func (t *task) notifyNext(value interface{}) {
	t.tracker.NotifyNext(value)
	t.tracker.NotifyCompleted()
}

func (t *task) notifyError(err error) {
	t.tracker.NotifyError(err)
	t.tracker.NotifyCompleted()
}

func (t *task) notifyCanceled(err error) {
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

// Subscribe allows the observer to listen for update to the current task
func (t *task) Subscribe(o Observer) io.Closer {
	return t.tracker.Subscribe(o)
}

// Unsubscribe allows the oberver to disconnect from updates to the current task
func (t *task) Unsubscribe(o Observer) {
	t.tracker.Unsubscribe(o)
}

// OnNext is called by a task that preceeds this current task
func (t *task) OnNext(next interface{}) {
}

func (t *task) OnError(err error) {
}

func (t *task) OnCanceled(err error) {
}

func (t *task) OnCompleted() {
	t.tracker.Close()

	// schedule this task irrespective of its status
	// we want the delegates with task parameters to handle the errors
	t.scheduler.Queue(t)
}

func (t *task) ContinueAction(continueAction ContinueAction) ObservableTask {
	f := func(t Task, state interface{}) (interface{}, error) {
		continueAction(t)
		return nil, nil
	}
	return t.ContinueErrFuncWith(f)
}

func (t *task) ContinueActionWith(continueActionWith ContinueActionWith) ObservableTask {
	f := func(t Task, state interface{}) (interface{}, error) {
		continueActionWith(t, state)
		return nil, nil
	}
	return t.ContinueErrFuncWith(f)
}

func (t *task) ContinueErrAction(continueErrAction ContinueErrAction) ObservableTask {
	f := func(t Task, state interface{}) (interface{}, error) {
		return nil, continueErrAction(t)
	}
	return t.ContinueErrFuncWith(f)
}

func (t *task) ContinueErrActionWith(continueErrActionWith ContinueErrActionWith) ObservableTask {
	f := func(t Task, state interface{}) (interface{}, error) {
		return nil, continueErrActionWith(t, state)
	}
	return t.ContinueErrFuncWith(f)
}

func (t *task) ContinueFunc(continueFunc ContinueFunc) ObservableTask {
	f := func(t Task, state interface{}) (interface{}, error) {
		return continueFunc(t), nil
	}
	return t.ContinueErrFuncWith(f)
}

func (t *task) ContinueFuncWith(continueFuncWith ContinueFuncWith) ObservableTask {
	f := func(t Task, state interface{}) (interface{}, error) {
		return continueFuncWith(t, state), nil
	}
	return t.ContinueErrFuncWith(f)
}

func (t *task) ContinueErrFunc(continueErrFunc ContinueErrFunc) ObservableTask {
	f := func(t Task, state interface{}) (interface{}, error) {
		return continueErrFunc(t)
	}
	return t.ContinueErrFuncWith(f)
}

func (t *task) ContinueErrFuncWith(continueErrFuncWith ContinueErrFuncWith) ObservableTask {
	errFuncWith := func(state interface{}) (interface{}, error) {
		return continueErrFuncWith(t, state)
	}
	continuation := new(errFuncWith)
	if t.context != nil {
		continuation.context = t.context
	}
	if t.scheduler != nil {
		continuation.scheduler = t.scheduler
	}

	// if the current task is already complete, immediately schedule the continuation
	// otherwise setup a subscription
	if t.IsCompleted() {
		continuation.scheduler.Queue(continuation)
	} else {
		t.Subscribe(continuation)
	}
	return continuation
}
