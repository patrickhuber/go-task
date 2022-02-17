package task

import (
	"context"
	"sync/atomic"
)

type whenTask struct {
	task
	tasks     []Task
	remaining int32
}

// WhenAny creates a task that completes when any task in the list completes
func WhenAny(tasks ...Task) Task {
	return when(1, tasks...)
}

// WhenAll creates a task that completes when all tasks in the list complete
func WhenAll(tasks ...Task) Task {
	return when(len(tasks), tasks...)
}

// When creates a task that completes when the limit of tasks complete
func when(limit int, tasks ...Task) Task {
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
		if t.IsCompleted() {
			when.OnCompleted()
			continue
		}
		t.Subscribe(when)
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
	defer t.tracker.Close()

	switch t.Status() {
	case StatusCanceled:
		t.notifyCanceled(t.Error())
	case StatusFaulted:
		t.notifyError(t.Error())
	case StatusSuccess:
		t.notifyNext(t.Result())
	}

	// notify that we are done by closing the channel
	defer close(t.doneCh)
}

func (t *whenTask) OnError(err error) {
}

func (t *whenTask) OnCanceled(err error) {
}
