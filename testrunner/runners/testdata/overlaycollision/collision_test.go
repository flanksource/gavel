package overlaycollision

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
)

func TestSuite(t *testing.T) {
	ginkgo.RunSpecs(t, "Overlay collision fixture")
}

var _ = ginkgo.Describe("package-level names", func() {
	ginkgo.It("coexist with the injected reporter", func() {
		_ = Report{}
		_ = Context{}
		_ = Labels{}
		if len(Names()) != 5 {
			ginkgo.Fail("package-level names are missing")
		}
	})
})
