package task_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickhuber/go-task"
)

var _ = Describe("WhenAny", func() {
	It("waits until one task completes", func() {
		scheduler := task.NewQueueScheduler()
		blocking := task.RunErrFuncWith(func(i interface{}) (interface{}, error) {
			ch := make(chan struct{})
			defer close(ch)
			select {
			case <-ch:
			case <-time.After(time.Second):
			}
			return nil, nil
		}, task.WithScheduler(scheduler))
		completed := task.Completed()
		t := task.WhenAny(blocking, completed)
		Expect(t.Wait()).To(BeNil())
	})
	It("returns immediately when no tasks", func() {
		t := task.WhenAny()
		Expect(t.Wait()).To(BeNil())
	})
	It("can process completed task", func() {
		t := task.WhenAny(task.Completed())
		Expect(t.Wait()).To(BeNil())
	})
})
