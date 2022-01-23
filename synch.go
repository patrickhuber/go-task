package task

type Synchronization interface {
	// Post invokes the delegate synchronously
	Post(delegate DelegateFunc) (interface{}, error)
	// Send invokes the delegate in a go routine
	Send(delegate DelegateFunc) (chan interface{}, chan error)
}

type synchronization struct {
}

func DefaultSynchronization() Synchronization {
	return &synchronization{}
}

func (s *synchronization) Post(delegate DelegateFunc) (interface{}, error) {
	return delegate()
}

func (s *synchronization) Send(delegate DelegateFunc) (chan interface{}, chan error) {
	errChan := make(chan error)
	resultChan := make(chan interface{})
	go func(delegate DelegateFunc) {
		// cleanup the channels when this goroutine ends
		defer close(errChan)
		defer close(resultChan)

		// execute the delegate
		result, err := delegate()

		// set the channels
		if err != nil {
			errChan <- err
		} else {
			resultChan <- result
		}
	}(delegate)
	return resultChan, errChan
}
