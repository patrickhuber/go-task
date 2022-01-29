package task

import "sync"

type List interface {
	RemoveAt(index int)
	Remove(value interface{})
	Append(value interface{})
	Set(index int, value interface{})
	Get(index int) interface{}
	Length() int
	Contains(value interface{}) bool
	IndexOf(value interface{}) int
	Apply(func(item interface{}))
}

type list struct {
	values []interface{}
}

func NewList(values []interface{}) List {
	return &list{
		values: values,
	}
}

func (l *list) Get(index int) interface{} {
	return l.values[index]
}

func (l *list) Set(index int, value interface{}) {
	l.values[index] = value
}

func (l *list) Append(value interface{}) {
	l.values = append(l.values, value)
}

func (l *list) RemoveAt(index int) {
	copy(l.values[index:], l.values[index+1:]) // Shift a[i+1:] left one index.
	l.values[len(l.values)-1] = nil            // Erase last element (write zero value).
	l.values = l.values[:len(l.values)-1]      // Truncate slice.
}

func (l *list) Remove(value interface{}) {
	index := l.IndexOf(value)
	l.RemoveAt(index)
}

func (l *list) Length() int {
	return len(l.values)
}

func (l *list) Contains(value interface{}) bool {
	return l.IndexOf(value) >= 0
}

func (l *list) IndexOf(value interface{}) int {
	index := -1
	for i, v := range l.values {
		if v == value {
			index = i
		}
	}
	return index
}

func (l *list) Apply(delegate func(item interface{})) {
	for _, item := range l.values {
		delegate(item)
	}
}

type concurrentList struct {
	innerList List
	mut       sync.RWMutex
}

func NewConcurrentList(values []interface{}) List {
	list := NewList(values)
	return &concurrentList{
		innerList: list,
	}
}

func (l *concurrentList) Get(index int) interface{} {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.innerList.Get(index)
}

func (l *concurrentList) Set(index int, value interface{}) {
	l.mut.Lock()
	defer l.mut.Unlock()
	l.innerList.Set(index, value)
}

func (l *concurrentList) Append(value interface{}) {
	l.mut.Lock()
	defer l.mut.Unlock()
	l.innerList.Append(value)
}

func (l *concurrentList) RemoveAt(index int) {
	l.mut.Lock()
	defer l.mut.Unlock()
	l.innerList.RemoveAt(index)
}

func (l *concurrentList) Remove(value interface{}) {
	index := l.IndexOf(value)
	l.RemoveAt(index)
}

func (l *concurrentList) Length() int {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.innerList.Length()
}

func (l *concurrentList) Contains(value interface{}) bool {
	// no mutex needed here because IndexOf takes care of the lock
	return l.IndexOf(value) >= 0
}

func (l *concurrentList) IndexOf(value interface{}) int {
	l.mut.RLock()
	defer l.mut.RUnlock()
	return l.innerList.IndexOf(value)
}

func (l *concurrentList) Apply(delegate func(item interface{})) {
	for i := 0; i < l.Length(); i++ {
		l.mut.Lock()
		value := l.innerList.Get(i)
		delegate(value)
		l.mut.Unlock()
	}
}
