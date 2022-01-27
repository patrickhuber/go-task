package task_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickhuber/go-task"
)

var _ = Describe("AggregateError", func() {
	Describe("AppendError", func() {
		When("one nil", func() {
			It("returns nil", func() {
				err := task.AppendError(nil)
				Expect(err).To(BeNil())
			})
		})
		When("empty", func() {
			It("returns nil", func() {
				err := task.AppendError()
				Expect(err).To(BeNil())
			})
		})
		When("all nil", func() {
			It("returns nil", func() {
				err := task.AppendError(nil)
				Expect(err).To(BeNil())
			})
		})
		When("one not nil", func() {
			It("returns aggregate", func() {
				errors := []error{nil, nil, nil, fmt.Errorf("not"), nil, nil}
				err := task.AppendError(errors...)
				Expect(err).ToNot(BeNil())
				Expect(len(err.Errors())).To(Equal(1))
			})
		})
		When("only aggregate", func() {
			It("returns aggregate", func() {
				agg := task.AppendError(fmt.Errorf("first"))
				err := task.AppendError(agg, fmt.Errorf("second"))
				Expect(err).ToNot(BeNil())
				Expect(len(err.Errors())).To(Equal(2))
			})
		})
	})
})
