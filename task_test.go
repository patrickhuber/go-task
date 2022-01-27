package task_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickhuber/go-task"
)

var _ = Describe("Task", func() {
	It("can return error", func() {
		t := task.RunErrFunc(func() (interface{}, error) {
			return nil, fmt.Errorf("this is an error")
		})
		err := t.Wait()
		Expect(err).ToNot(BeNil())
		result := t.Result()
		Expect(result).To(BeNil())
	})
	It("can return result", func() {
		t := task.RunFunc(func() interface{} {
			return 1
		})
		err := t.Wait()
		Expect(err).To(BeNil())
		result := t.Result()
		Expect(result).ToNot(BeNil())
		Expect(result).To(Equal(1))
	})
	It("can timeout", func() {
		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond)
		t := task.RunAction(func() {
			ch := make(chan struct{})
			<-ch
		}, task.WithContext(ctx))
		Expect(t.Wait()).ToNot(BeNil())
	})
	It("can cancel", func() {
		ctx, cancel := context.WithCancel(context.Background())
		t := task.RunAction(func() {
			ch := make(chan struct{})
			<-ch
		}, task.WithContext(ctx))
		cancel()
		Expect(t.Wait()).ToNot(BeNil())
	})
	Describe("FromResult", func() {
		It("is completed", func() {
			expected := 1
			t := task.FromResult(expected)
			Expect(t.IsCompleted()).To(BeTrue(), "task is not complete")
			Expect(t.Wait()).To(BeNil())
			Expect(t.Result()).To(Equal(expected))
		})
	})
	Describe("FromError", func() {
		It("is completed", func() {
			expected := fmt.Errorf("test")
			t := task.FromError(expected)
			Expect(t.IsCompleted()).To(BeTrue(), "task is not complete")
			Expect(t.Wait()).ToNot(BeNil())
			Expect(t.Error()).ToNot(BeNil())
		})
	})
	Describe("Complete", func() {
		It("is complete", func() {
			t := task.Completed()
			Expect(t.IsCompleted()).To(BeTrue())
			Expect(t.Wait()).To(BeNil())
		})
	})
	Describe("Action", func() {
		It("can execute", func() {
			t := task.RunAction(func() {})
			Expect(t.Wait()).To(BeNil())
			Expect(t.Result()).To(BeNil())
		})
	})
	Describe("ActionWith", func() {
		It("can pass state", func() {
			expected := 1
			t := task.RunActionWith(func(state interface{}) {
				Expect(state).ToNot(BeNil())
				Expect(state).To(Equal(expected))
			}, task.WithState(expected))
			Expect(t.Wait()).To(BeNil())
			Expect(t.Result()).To(BeNil())
		})
	})
})
