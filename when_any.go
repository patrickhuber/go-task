package task

// WhenAny sets up event notifications
func WhenAny(tasks ...ObservableTask) ObservableTask {
	return When(1, tasks...)
}
