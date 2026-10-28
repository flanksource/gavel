package overlaymissing

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
)

func TestWithoutSuite(t *testing.T) {
	t.Logf("Ginkgo is imported without a RunSpecs entry point: seed=%d", ginkgo.GinkgoRandomSeed())
}
