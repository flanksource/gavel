package todos

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTodos(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TODO Suite")
}
