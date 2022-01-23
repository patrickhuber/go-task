package task

import (
	"context"
	"io"
)

type Task interface {
	Wait() error
	Result() interface{}
	Error() error
	IsCompleted() bool
}

type Observer interface {
	OnNext(interface{})
	OnCompleted()
	OnError(error)
}

type Observable interface {
	Subscribe(Observer) io.Closer
}

type task struct {
	errChan         chan error
	resultChan      chan interface{}
	result          interface{}
	err             error
	completed       bool
	context         context.Context
	synchronization Synchronization
}

type RunOption func(t *task)

func WithContext(ctx context.Context) RunOption {
	return func(t *task) {
		t.context = ctx
	}
}

func WithSync(s Synchronization) RunOption {
	return func(t *task) {
		t.synchronization = s
	}
}

func (t *task) Wait() error {
	// result is cached? return it
	if t.IsCompleted() {
		return t.err
	}

	select {
	case res := <-t.resultChan:
		t.result = res
		t.completed = true
		return nil
	case err := <-t.errChan:
		t.err = err
		t.completed = true
		return err
	case <-t.context.Done():
		t.err = t.context.Err()
		t.completed = true
		return t.err
	}
}

func (t *task) Result() interface{} {
	return t.result
}

func (t *task) Error() error {
	return t.err
}

func (t *task) IsCompleted() bool {
	return t.completed
}

func FromResult(result interface{}) Task {
	return &task{
		result:    result,
		completed: true,
	}
}

func FromError(err error) Task {
	return &task{
		err:       err,
		completed: true,
	}
}

func Run(delegate DelegateFunc, options ...RunOption) Task {
	// create the task
	t := &task{
		context:         context.TODO(),
		completed:       false,
		synchronization: DefaultSynchronization(),
	}

	// apply operations
	for _, opt := range options {
		opt(t)
	}

	t.resultChan, t.errChan = t.synchronization.Send(delegate)
	return t
}
