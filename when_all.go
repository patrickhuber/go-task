package task

func WhenAll(tasks ...ObservableTask) ObservableTask {
	return When(len(tasks), tasks...)
}
