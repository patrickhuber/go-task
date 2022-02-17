package task_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/patrickhuber/go-task"
)

var _ = Describe("New", func() {
	It("can create new Action", func() {
		t := task.NewAction(func() {
		})
		Expect(t).ToNot(BeNil())
		Expect(t.Status()).To(Equal(task.StatusCreated))
	})
	It("can create new ActionWith", func() {
		t := task.NewActionWith(func(state interface{}) {
		}, task.WithState(1))
		Expect(t).ToNot(BeNil())
		Expect(t.Status()).To(Equal(task.StatusCreated))
	})
	It("can create new ErrAction", func() {
		t := task.NewErrAction(func() error {
			return nil
		})
		Expect(t).ToNot(BeNil())
		Expect(t.Status()).To(Equal(task.StatusCreated))
	})
	It("can create new ErrActionWith", func() {
		t := task.NewErrActionWith(func(state interface{}) error {
			return nil
		}, task.WithState(1))
		Expect(t).ToNot(BeNil())
		Expect(t.Status()).To(Equal(task.StatusCreated))
	})
	It("can create new Func", func() {
		t := task.NewFunc(func() interface{} {
			return nil
		})
		Expect(t).ToNot(BeNil())
		Expect(t.Status()).To(Equal(task.StatusCreated))
	})
	It("can create new FuncWith", func() {
		t := task.NewFuncWith(func(state interface{}) interface{} {
			return nil
		}, task.WithState(1))
		Expect(t).ToNot(BeNil())
		Expect(t.Status()).To(Equal(task.StatusCreated))
	})
	It("can create new ErrFunc", func() {
		t := task.NewErrFunc(func() (interface{}, error) {
			return nil, nil
		})
		Expect(t).ToNot(BeNil())
		Expect(t.Status()).To(Equal(task.StatusCreated))
	})
	It("can create new ErrFuncWith", func() {
		t := task.NewErrFuncWith(func(state interface{}) (interface{}, error) {
			return nil, nil
		}, task.WithState(1))
		Expect(t).ToNot(BeNil())
		Expect(t.Status()).To(Equal(task.StatusCreated))
	})
})
