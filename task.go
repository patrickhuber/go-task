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
	status    TaskStatus
	result    interface{}
	err       error
	doneCh    chan struct{}
	context   context.Context
	scheduler Scheduler
	delegate  DelegateFunc
	state     interface{}
	observers []Observer
}

func new(delegate DelegateFunc) *task {
	return &task{
		context:   context.TODO(),
		status:    StatusCreated,
		scheduler: DefaultScheduler(),
		delegate:  delegate,
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
	result, err := t.delegate(t.state)
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

func Completed() Task {
	return &task{
		status: StatusSuccess,
	}
}

func FromResult(result interface{}) Task {
	return &task{
		result: result,
		status: StatusSuccess,
	}
}

func FromError(err error) Task {
	return &task{
		err:    err,
		status: StatusFaulted,
	}
}

func Run(delegate DelegateFunc, options ...RunOption) ObservableTask {
	// create the task
	t := new(delegate)

	// apply operations
	for _, opt := range options {
		opt(t)
	}
	t.scheduler.Queue(t)
	return t
}
