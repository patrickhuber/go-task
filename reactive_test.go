package task_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickhuber/go-task"
)

type testObserver struct {
	errorCount     int
	nextCount      int
	completedCount int
	canceledCount  int
	subscribers    []task.Observer
}

func NewTestObserver() *testObserver {
	return &testObserver{
		subscribers: []task.Observer{},
	}
}

func (p *testObserver) OnNext(interface{}) {
	p.nextCount++
}

func (p *testObserver) OnCompleted() {
	p.completedCount++
}

func (p *testObserver) OnCanceled(error) {
	p.canceledCount++
}

func (p *testObserver) OnError(error) {
	p.errorCount++
}

var _ = Describe("Subscriber", func() {
	var (
		observer *testObserver
		tracker  task.Tracker
	)
	BeforeEach(func() {
		tracker = task.NewTracker()
		observer = NewTestObserver()
	})
	It("can publish next", func() {
		subscription := tracker.Subscribe(observer)
		defer subscription.Close()
		tracker.NotifyNext(1)
		Expect(observer.nextCount).To(Equal(1))
	})
	It("can publish error", func() {
		subscription := tracker.Subscribe(observer)
		defer subscription.Close()
		tracker.NotifyError(fmt.Errorf("error"))
		Expect(observer.errorCount).To(Equal(1))
	})
	It("can publish canceled", func() {
		subscription := tracker.Subscribe(observer)
		defer subscription.Close()
		tracker.NotifyCanceled(fmt.Errorf("error"))
		Expect(observer.canceledCount).To(Equal(1))
	})
	It("can publish completed", func() {
		subscription := tracker.Subscribe(observer)
		defer subscription.Close()
		tracker.NotifyCompleted()
		Expect(observer.completedCount).To(Equal(1))
	})
	It("can unsubscribe", func() {
		subscription := tracker.Subscribe(observer)
		tracker.NotifyCompleted()
		Expect(observer.completedCount).To(Equal(1))
		Expect(subscription.Close()).To(BeNil())
		tracker.NotifyCompleted()
		Expect(observer.completedCount).To(Equal(1))
	})
})
