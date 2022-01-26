package task

import "io"

type Observer interface {
	OnNext(interface{})
	OnCompleted()
	OnCanceled(error)
	OnError(error)
}

type Observable interface {
	Subscribe(Observer) io.Closer
}

type subscription struct {
	observer  Observer
	observers []Observer
}

func newSubscription(observer Observer, observers []Observer) io.Closer {
	return &subscription{
		observer:  observer,
		observers: observers,
	}
}

func (s *subscription) Close() error {
	s.remove(s.observer)
	return nil
}

func (s *subscription) remove(observer Observer) {
	if observer == nil {
		return
	}
	index := s.index(observer)
	if index < 0 {
		return
	}
	s.observers[index] = s.observers[len(s.observers)-1]
	s.observers[len(s.observers)-1] = nil
	s.observers = s.observers[:len(s.observers)-1]
}

func (s *subscription) index(observer Observer) int {
	for i, o := range s.observers {
		if o == observer {
			return i
		}
	}
	return -1
}
