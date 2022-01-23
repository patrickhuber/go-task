package task_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGoTask(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoTask Suite")
}
