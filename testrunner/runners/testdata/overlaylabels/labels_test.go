package overlaylabels

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLabels(t *testing.T) {
	RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Labels overlay")
}

var _ = ginkgo.Describe("Labels", func() {
	ginkgo.It("retains the package's labels type", func() {
		Expect(Labels{"app": "example"}).To(HaveKeyWithValue("app", "example"))
	})
})
