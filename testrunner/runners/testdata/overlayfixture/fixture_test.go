package overlayfixture

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOrdinary(t *testing.T) {
	if 2+2 != 4 {
		t.Fatal("arithmetic failed")
	}
}

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Overlay fixture")
}

var _ = Describe("container", Label("inherited"), func() {
	It("reports metadata", Label("leaf"), func() {
		AddReportEntry("resource", map[string]string{"kind": "Widget"})
		Expect(true).To(BeTrue())
	})

	// Succinct output ("•" without a newline) must not swallow the first metadata frame.
	It("passes without report entries", func() {
		Expect(true).To(BeTrue())
	})
})
