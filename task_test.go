package task_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickhuber/task"
)

var _ = Describe("Task", func() {
	It("can return error", func() {
		t := task.Run(func() (interface{}, error) {
			return nil, fmt.Errorf("this is an error")
		})
		err := t.Wait()
		Expect(err).ToNot(BeNil())
		result := t.Result()
		Expect(result).To(BeNil())
	})
	It("can return result", func() {
		t := task.Run(func() (interface{}, error) {
			return 1, nil
		})
		err := t.Wait()
		Expect(err).To(BeNil())
		result := t.Result()
		Expect(result).ToNot(BeNil())
	})
	It("can timeout", func() {
		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond)
		t := task.Run(func() (interface{}, error) {
			<-time.After(time.Second * 10)
			return nil, nil
		}, task.WithContext(ctx))
		Expect(t.Wait()).ToNot(BeNil())
	})
	It("can cancel", func() {
		ctx, cancel := context.WithCancel(context.Background())
		t := task.Run(func() (interface{}, error) {
			<-time.After(time.Second * 10)
			return nil, nil
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
	Describe("WhenAll", func() {})
	Describe("WhenAny", func() {})
})
