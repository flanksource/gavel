package shims

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestShims(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linter shims")
}
