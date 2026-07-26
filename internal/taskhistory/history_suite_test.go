package taskhistory_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTaskHistory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Task History Suite")
}
