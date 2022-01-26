package task

import "context"

type whenAnyTask struct {
	task
	tasks []Observable
}

// WhenAny sets up event notifications
func WhenAny(tasks []Observable) Task {
	return &whenAnyTask{
		task: task{
			doneCh:  make(chan struct{}, 1),
			context: context.TODO(),
		},
		tasks: tasks,
	}
}

func (t *whenAnyTask) OnNext(value interface{}) {

}

func (t *whenAnyTask) OnCompleted() {
	t.MarkComplete()
}

func (t *whenAnyTask) OnError(err error) {
}

func (t *whenAnyTask) OnCanceled() {
}

func (t *whenAnyTask) MarkComplete() {
	if !t.IsCompleted() {
		t.doneCh <- struct{}{}
	}
}
