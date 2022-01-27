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
	Unsubscribe(Observer)
}

type subscription struct {
	observer   Observer
	observable Observable
}

func NewSubscription(observer Observer, observable Observable) io.Closer {
	return &subscription{
		observer:   observer,
		observable: observable,
	}
}

func (s *subscription) Close() error {
	s.observable.Unsubscribe(s.observer)
	return nil
}

type tracker struct {
	observers []Observer
}

type Tracker interface {
	Observable
	NotifyError(error)
	NotifyNext(interface{})
	NotifyCompleted()
	NotifyCanceled(error)
}

func NewTracker() Tracker {
	return &tracker{
		observers: []Observer{},
	}
}

func (t *tracker) Subscribe(observer Observer) io.Closer {
	contains := t.indexOf(observer) >= 0
	if !contains {
		t.observers = append(t.observers, observer)
	}
	return NewSubscription(observer, t)
}

func (t *tracker) Unsubscribe(observer Observer) {
	index := t.indexOf(observer)
	if index < 0 {
		return
	}
	t.observers[index] = t.observers[len(t.observers)-1]
	t.observers[len(t.observers)-1] = nil
	t.observers = t.observers[:len(t.observers)-1]
}

func (t *tracker) indexOf(observer Observer) int {
	index := -1
	for i, o := range t.observers {
		if o == observer {
			index = i
		}
	}
	return index
}

func (t *tracker) NotifyError(err error) {
	for _, o := range t.observers {
		o.OnError(err)
	}
}

func (t *tracker) NotifyNext(next interface{}) {
	for _, o := range t.observers {
		o.OnNext(next)
	}
}

func (t *tracker) NotifyCompleted() {
	for _, o := range t.observers {
		o.OnCompleted()
	}
}

func (t *tracker) NotifyCanceled(err error) {
	for _, o := range t.observers {
		o.OnCanceled(err)
	}
}
