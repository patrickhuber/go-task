package task

type Scheduler interface {
	Queue(t Task)
}

type scheduler struct {
}

func DefaultScheduler() Scheduler {
	return &scheduler{}
}

func NewScheduler() Scheduler {
	return &scheduler{}
}

func (s *scheduler) Queue(t Task) {
	// do not schedule completed tasks
	if t.IsCompleted() {
		return
	}
	go func(t Task) {
		t.Start()
	}(t)
	go func(t Task) {
		t.Wait()
	}(t)
}
