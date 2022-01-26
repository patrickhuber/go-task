package task

import (
	"context"
	"io"
	"sync/atomic"
)

type whenAllTask struct {
	task
	tasks         []ObservableTask
	remaining     int32
	subscriptions []io.Closer
}

func WhenAll(tasks []ObservableTask) ObservableTask {
	whenAll := &whenAllTask{
		tasks:     tasks,
		remaining: int32(len(tasks)),
		task: task{
			context: context.TODO(),
			// use a buffered channel to avoid blocking caller
			doneCh: make(chan struct{}, 1),
		},
	}

	for _, t := range tasks {
		subscription := t.Subscribe(whenAll)
		whenAll.subscriptions = append(whenAll.subscriptions, subscription)
	}
	return whenAll
}

func (t *whenAllTask) IsCompleted() bool {
	return t.remaining == 0
}

func (t *whenAllTask) Execute() {
	// do nothing because this is more of a promise
	// than a task
}

func (t *whenAllTask) OnNext(value interface{}) {
}

func (t *whenAllTask) OnCompleted() {
	if atomic.AddInt32(&t.remaining, -1) != 0 {
		return
	}

	// start in a success state
	t.status = StatusSuccess

	// all tasks are completed, process them
	for i := 0; i < len(t.tasks); i++ {
		tsk := t.tasks[i]
		if tsk.IsFaulted() {
			t.status = StatusFaulted
			if tsk.Error() != nil {
				t.err = AppendError(t.err, tsk.Error())
			}
		} else if tsk.IsCanceled() {
			t.status = StatusCanceled
			if tsk.Error() != nil {
				t.err = AppendError(t.err, tsk.Error())
			}
		}
	}

	// clear all subscriptions
	for _, s := range t.subscriptions {
		s.Close()
	}
	t.subscriptions = nil

	switch t.status {
	case StatusCanceled:
		t.NotifyCanceled(t.err)
	case StatusFaulted:
		t.NotifyError(t.err)
	case StatusSuccess:
		t.NotifyNext(t.result)
	}

	// notify that we are done
	t.doneCh <- struct{}{}
}

func (t *whenAllTask) OnError(err error) {
}

func (t *whenAllTask) OnCanceled(err error) {
}
