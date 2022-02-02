package task

import (
	"io"

	"github.com/patrickhuber/go-collections/list"
	concurrent_list "github.com/patrickhuber/go-collections/concurrent/list"
)

type Observer interface {
	OnNext(interface{})
	OnCompleted()
	OnCanceled(error)
	OnError(error)
}

type observer struct {
	onNext      func(interface{})
	onCompleted func()
	onCanceled  func(error)
	onError     func(error)
}

func NewObserver(
	onNext func(interface{}),
	onCompleted func(),
	onCanceled func(error),
	onError func(error)) Observer {
	return &observer{
		onNext:      onNext,
		onCompleted: onCompleted,
		onCanceled:  onCanceled,
		onError:     onError,
	}
}

func (o *observer) OnNext(next interface{}) {
	if o.onNext != nil {
		o.onNext(next)
	}
}

func (o *observer) OnCompleted() {
	if o.onCompleted != nil {
		o.onCompleted()
	}
}

func (o *observer) OnCanceled(err error) {
	if o.onCanceled != nil {
		o.onCanceled(err)
	}
}

func (o *observer) OnError(err error) {
	if o.onError != nil {
		o.onError(err)
	}
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
	observers list.List
}

type Tracker interface {
	Observable
	io.Closer
	NotifyError(error)
	NotifyNext(interface{})
	NotifyCompleted()
	NotifyCanceled(error)
}

func NewTracker() Tracker {
	return &tracker{
		observers: concurrent_list.New(),
	}
}

func (t *tracker) Subscribe(observer Observer) io.Closer {
	if !t.observers.Contains(observer) {
		t.observers.Append(observer)
	}
	return NewSubscription(observer, t)
}

func (t *tracker) Unsubscribe(observer Observer) {
	if !t.observers.Contains(observer) {
		return
	}
	t.observers.Remove(observer)
}

func (t *tracker) NotifyError(err error) {
	t.notify(func(o Observer) {
		o.OnError(err)
	})
}

func (t *tracker) NotifyNext(next interface{}) {
	t.notify(func(o Observer) {
		o.OnNext(next)
	})
}

func (t *tracker) NotifyCompleted() {
	t.notify(func(o Observer) {
		o.OnCompleted()
	})
}

func (t *tracker) NotifyCanceled(err error) {
	t.notify(func(o Observer) {
		o.OnCanceled(err)
	})
}

func (t *tracker) notify(action func(o Observer)) {
	t.observers.ForEach(func(item interface{}) {
		if o, exists := item.(Observer); exists {
			action(o)
		}
	})
}

func (t *tracker) Close() error {
	t.observers.Clear()
	return nil
}
