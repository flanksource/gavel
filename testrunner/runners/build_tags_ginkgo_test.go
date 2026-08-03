package runners

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("build tag mappings", func() {
	It("maps tags to go test", func() {
		Expect(NewGoTest(GinkgoT().TempDir()).BuildTagsArgs([]string{"integration", "postgres"})).To(
			Equal([]string{"-tags=integration,postgres"}),
		)
	})

	It("maps tags to ginkgo", func() {
		Expect(NewGinkgo(GinkgoT().TempDir()).BuildTagsArgs([]string{"integration", "postgres"})).To(
			Equal([]string{"--tags=integration,postgres"}),
		)
	})
})
