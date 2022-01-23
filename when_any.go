package task

type whenAnyTask struct {
	task
	tasks []Observable
}

// WhenAny sets up event notifications
func WhenAny(tasks []Observable) Task {
	return &whenAnyTask{
		tasks: tasks,
	}
}

func (t *whenAnyTask) OnNext(value interface{}) {
	t.result = value
	t.completed = true
	t.resultChan <- value
}

func (t *whenAnyTask) OnCompleted() {
	
}

func (t *whenAnyTask) OnError(err error) {
	t.err = err
	t.completed = true
	t.errChan <- err
}
