package task_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickhuber/task"
)

var _ = Describe("WhenAll", func() {
	It("waits until all tasks complete", func() {
		itemCount := 3
		tasks := []task.ObservableTask{}
		scheduler := task.NewQueueScheduler()
		for i := 0; i < itemCount; i++ {
			t := task.Run(func(state interface{}) (interface{}, error) {
				i := state.(int)
				return i, nil
			}, task.WithScheduler(scheduler),
				task.WithState(i))
			tasks = append(tasks, t)
		}

		t := task.WhenAll(tasks...)

		for i := 0; i < itemCount; i++ {
			Expect(scheduler.Dequeue()).To(BeTrue())
		}
		Expect(t.Wait()).To(BeNil())
	})
	It("returns cancel err when one task is canceled", func() {

		scheduler := task.NewQueueScheduler()
		normal := task.Run(func(state interface{}) (interface{}, error) {
			i := state.(int)
			return i, nil
		}, task.WithScheduler(scheduler),
			task.WithState(10))

		cancel := task.Run(func(interface{}) (interface{}, error) {
			ch := make(chan struct{})
			<-ch
			return 10, nil
		}, task.WithScheduler(scheduler),
			task.WithTimeout(time.Millisecond*10))

		tasks := []task.ObservableTask{normal, cancel}
		t := task.WhenAll(tasks...)
		for i := 0; i < len(tasks); i++ {
			Expect(scheduler.Dequeue()).To(BeTrue())
		}
		Expect(t.Wait()).ToNot(BeNil())
	})
	It("returns immediately when no tasks", func() {
		t := task.WhenAll()
		Expect(t.Wait()).To(BeNil())
	})
	It("can process completed task", func() {
		t := task.WhenAll(task.Completed())

		Expect(t.Wait()).To(BeNil())
	})
})
