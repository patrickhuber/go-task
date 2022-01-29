package task_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/patrickhuber/go-task"
)

var _ = Describe("Continuation", func() {
	It("can subscribe", func() {
		err := task.RunAction(func() {
		}).ContinueAction(func(t task.Task) {
			Expect(t).ToNot(BeNil())
			Expect(t.IsCompleted()).To(BeTrue())
		}).Wait()
		Expect(err).To(BeNil())
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
})
