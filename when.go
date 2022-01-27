package task

import (
	"context"
	"io"
	"sync/atomic"
)

type whenTask struct {
	task
	tasks         []ObservableTask
	remaining     int32
	subscriptions []io.Closer
}

// WhenAny creates a task that completes when any task in the list completes
func WhenAny(tasks ...ObservableTask) ObservableTask {
	return When(1, tasks...)
}

// WhenAll creates a task that completes when all tasks in the list complete
func WhenAll(tasks ...ObservableTask) ObservableTask {
	return When(len(tasks), tasks...)
}

func When(limit int, tasks ...ObservableTask) ObservableTask {
	if len(tasks) == 0 {
		return Completed()
	}
	when := &whenTask{
		tasks:     tasks,
		remaining: int32(limit),
		task: task{
			tracker: NewTracker(),
			context: context.TODO(),
			// use a buffered channel to avoid blocking caller
			doneCh: make(chan struct{}, 1),
		},
	}

	for _, t := range tasks {

		// bypass the subscription if the task is completed
		// is there a race condition when tasks are scheduled immediately?
		if t.IsCompleted() {
			when.OnCompleted()
			continue
		}
		subscription := t.Subscribe(when)
		when.subscriptions = append(when.subscriptions, subscription)
	}

	return when
}

func (t *whenTask) IsCompleted() bool {
	return t.remaining == 0
}

func (t *whenTask) Execute() {
	// do nothing because this is more of a promise
	// than a task
}

func (t *whenTask) OnNext(value interface{}) {
}

func (t *whenTask) OnCompleted() {
	if atomic.AddInt32(&t.remaining, -1) != 0 {
		return
	}

	// start in a success state
	t.setStatus(StatusSuccess)

	// the remaining tasks are complete, process them
	// for WhenAll this is all tasks
	// for WhenAny this is any completed task
	for i := 0; i < len(t.tasks); i++ {
		tsk := t.tasks[i]
		if tsk.IsFaulted() {
			t.setStatus(StatusFaulted)
			if tsk.Error() != nil {
				t.setError(AppendError(t.Error(), tsk.Error()))
			}
		} else if tsk.IsCanceled() {
			t.setStatus(StatusCanceled)
			if tsk.Error() != nil {
				t.setError(AppendError(t.Error(), tsk.Error()))
			}
		}
	}

	// clear all subscriptions
	for _, s := range t.subscriptions {
		s.Close()
	}
	t.subscriptions = nil

	switch t.Status() {
	case StatusCanceled:
		t.NotifyCanceled(t.Error())
	case StatusFaulted:
		t.NotifyError(t.Error())
	case StatusSuccess:
		t.NotifyNext(t.Result())
	}

	// notify that we are done
	t.doneCh <- struct{}{}
}

func (t *whenTask) OnError(err error) {
}

func (t *whenTask) OnCanceled(err error) {
}
