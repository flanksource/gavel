package todoprojection

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTodoProjection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TODO Projection Suite")
}
