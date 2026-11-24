package branchrebase

import (
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Rebase commit IDs", func() {
	ginkgo.DescribeTable("accepts complete hexadecimal object IDs", func(sha string) {
		Expect(validateCommitID(sha)).To(Succeed())
	},
		ginkgo.Entry("SHA-1", strings.Repeat("a1", 20)),
		ginkgo.Entry("SHA-256", strings.Repeat("a1", 32)),
		ginkgo.Entry("uppercase", strings.Repeat("AB", 20)),
	)

	ginkgo.DescribeTable("rejects values that are not complete hexadecimal object IDs", func(sha string) {
		Expect(validateCommitID(sha)).NotTo(Succeed())
	},
		ginkgo.Entry("empty", ""),
		ginkgo.Entry("option", "--exec=unexpected"),
		ginkgo.Entry("abbreviated", strings.Repeat("a", 39)),
		ginkgo.Entry("extended SHA-1", strings.Repeat("a", 41)),
		ginkgo.Entry("short SHA-256", strings.Repeat("a", 63)),
		ginkgo.Entry("extended SHA-256", strings.Repeat("a", 65)),
		ginkgo.Entry("non-hexadecimal", strings.Repeat("z", 40)),
		ginkgo.Entry("revision expression", strings.Repeat("a", 40)+"^{commit}"),
		ginkgo.Entry("newline suffix", strings.Repeat("a", 40)+"\n"),
		ginkgo.Entry("space prefix", " "+strings.Repeat("a", 40)),
	)
})
