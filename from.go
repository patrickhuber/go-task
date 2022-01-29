package task

// Completed returns a completed task in the StatusSuccess state
func Completed() ObservableTask {
	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}
	return &task{
		status: StatusSuccess,
		doneCh: doneCh,
	}
}

// FromResult returns a completed task in the StatusSuccess state with the given result
func FromResult(result interface{}) ObservableTask {
	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}
	return &task{
		result: result,
		status: StatusSuccess,
		doneCh: doneCh,
	}
}

// FromError returns a completed task in StatusFaulted state with the given error
func FromError(err error) ObservableTask {
	doneCh := make(chan struct{}, 1)
	doneCh <- struct{}{}
	return &task{
		err:    err,
		status: StatusFaulted,
		doneCh: doneCh,
	}
}
