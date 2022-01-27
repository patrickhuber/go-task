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
	go func(t Task) {
		t.Execute()
	}(t)
	go func(t Task) {
		t.Wait()
	}(t)
}
