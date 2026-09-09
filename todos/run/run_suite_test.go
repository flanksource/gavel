package run_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTodosRun(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "todos/run")
}
