package task

// Completed returns a completed task in the StatusSuccess state
func Completed() Task {
	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}
	close(doneCh)
	return &task{
		status: StatusSuccess,
		doneCh: doneCh,
	}
}

// FromResult returns a completed task in the StatusSuccess state with the given result
func FromResult(result interface{}) Task {
	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}
	close(doneCh)
	return &task{
		result: result,
		status: StatusSuccess,
		doneCh: doneCh,
	}
}

// FromError returns a completed task in StatusFaulted state with the given error
func FromError(err error) Task {
	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}
	close(doneCh)
	return &task{
		err:    err,
		status: StatusFaulted,
		doneCh: doneCh,
	}
}

// FromAction creates an unstarted task from the given action
func FromAction(action Action) Task {
	return new(func(i interface{}) (interface{}, error) {
		action()
		return nil, nil
	})
}
