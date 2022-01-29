package task_test

import (
	"fmt"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/patrickhuber/go-task"
)

var _ = Describe("Continuation", func() {
	It("can subscribe", func() {
		count := int32(0)
		t := task.RunAction(func() {
			atomic.AddInt32(&count, 1)
		})
		cont := t.ContinueAction(func(t task.Task) {
			atomic.AddInt32(&count, 1)
		})
		err := cont.Wait()
		Expect(err).To(BeNil())
		Expect(count).To(Equal(int32(2)))
	})
	It("can forward task error", func() {
		err := task.RunErrAction(func() error {
			return fmt.Errorf("this is an error")
		}).ContinueErrAction(func(t task.Task) error {
			Expect(t).ToNot(BeNil())
			Expect(t.Error()).ToNot(BeNil())
			return t.Error()
		})
		Expect(err).ToNot(BeNil())
	})
	When("task complete", func() {
		It("queues task immediately", func() {
			count := 0
			t := task.Completed()
			c := t.ContinueAction(func(t task.Task) {
				count++
			})
			Expect(c.Wait()).To(BeNil())
			Expect(count).To(Equal(1))
		})
	})
})
