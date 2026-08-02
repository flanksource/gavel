package testrunner

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTestRunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Runner Suite")
}
