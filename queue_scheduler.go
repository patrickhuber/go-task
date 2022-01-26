package task

type queueScheduler struct {
	tasks []Task
}

type QueueScheduler interface {
	Scheduler
	Dequeue() bool
}

func NewQueueScheduler() QueueScheduler {
	return &queueScheduler{
		tasks: []Task{},
	}
}

func (s *queueScheduler) Queue(t Task) {
	s.tasks = append(s.tasks, t)
}

func (s *queueScheduler) Dequeue() bool {
	if len(s.tasks) == 0 {
		return false
	}
	t := s.tasks[len(s.tasks)-1]
	s.tasks[len(s.tasks)-1] = nil
	s.tasks = s.tasks[:len(s.tasks)-1]
	go func(t Task) {
		t.Execute()
	}(t)
	go func(t Task) {
		t.Wait()
	}(t)
	return true
}

func (s queueScheduler) DequeueAll() {
	for {
		if !s.Dequeue() {
			break
		}
	}
}
