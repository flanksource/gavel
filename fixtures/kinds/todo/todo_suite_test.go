package todo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTodoKinds(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Todo Fixture Kinds Suite")
}
