package prwatch

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPRWatch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PR Watch Suite")
}
