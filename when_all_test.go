package task_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickhuber/go-task"
)

var _ = Describe("WhenAll", func() {
	It("waits until all tasks complete", func() {
		itemCount := 3
		tasks := []task.ObservableTask{}
		scheduler := task.NewQueueScheduler()
		for i := 0; i < itemCount; i++ {
			t := task.RunFuncWith(func(state interface{}) interface{} {
				i := state.(int)
				return i
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
		normal := task.RunFuncWith(func(state interface{}) interface{} {
			i := state.(int)
			return i
		}, task.WithScheduler(scheduler),
			task.WithState(10))

		cancel := task.RunAction(func() {
			ch := make(chan struct{})
			<-ch
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
	It("returns aggregate errors", func() {
		tasks := []task.ObservableTask{}
		for i := 0; i < 3; i++ {
			t := task.FromError(fmt.Errorf("error %d", i))
			tasks = append(tasks, t)
		}
		t := task.WhenAll(tasks...)
		err := t.Wait()
		Expect(err).ToNot(BeNil())

		aggregate, ok := err.(task.AggregateError)
		Expect(ok).To(BeTrue())
		Expect(aggregate).ToNot(BeNil())
		Expect(len(aggregate.Errors())).To(Equal(3))
	})
})
