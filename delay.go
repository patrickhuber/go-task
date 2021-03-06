package task

import "time"

// Delay calls time.Sleep with the given duration
func Delay(duration time.Duration, options ...RunOption) Task {
	return RunAction(func() {
		time.Sleep(duration)
	}, options...)
}
