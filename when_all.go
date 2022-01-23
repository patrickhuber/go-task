package task

type whenAllTask struct {
	task
	tasks           []Observable
	completionCount int
}

func WhenAll(tasks []Observable) Task {
	whenAll := &whenAllTask{
		tasks:           tasks,
		completionCount: 0,
	}
	for _, t := range tasks {
		t.Subscribe(whenAll)
	}
	return whenAll
}

func (t *whenAllTask) OnNext(value interface{}) {
	if t.completed || t.completionCount == len(t.tasks) {
		return
	}
	t.completionCount++
	if t.completionCount != len(t.tasks) {
		return
	}
	t.completed = true
	t.resultChan <- value
}

func (t *whenAllTask) OnCompleted() {
}

func (t *whenAllTask) OnError(err error) {
	t.err = err
	t.completed = true
	t.errChan <- err
}
