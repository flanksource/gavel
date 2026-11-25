package testrunner

import (
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

var _ = Describe("filterFrameworks", func() {
	detected := func() map[Framework]*runners.TestOptions {
		return map[Framework]*runners.TestOptions{
			parsers.GoTest: {Cwd: "/repo"}, parsers.Ginkgo: {Cwd: "/repo"},
			parsers.Jest: {Cwd: "/repo"}, parsers.Vitest: {Cwd: "/repo"},
		}
	}

	It("narrows to the requested framework and keeps its detected options", func() {
		all := detected()
		got, err := filterFrameworks(all, []string{"jest"})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(map[Framework]*runners.TestOptions{parsers.Jest: all[parsers.Jest]}))
	})

	DescribeTable("resolves go aliases", func(alias string) {
		got, err := filterFrameworks(detected(), []string{alias})
		Expect(err).NotTo(HaveOccurred())
		Expect(lo.Keys(got)).To(Equal([]Framework{parsers.GoTest}))
	}, Entry("go", "go"), Entry("gotest", "gotest"), Entry("go-test", "go-test"), Entry("go test", "go test"))

	It("dedupes a framework requested twice", func() {
		got, err := filterFrameworks(detected(), []string{"jest", "jest"})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(HaveLen(1))
	})

	It("rejects an unknown framework, listing the known ones", func() {
		_, err := filterFrameworks(detected(), []string{"jst"})
		Expect(err).To(MatchError(And(ContainSubstring("unknown framework"), ContainSubstring("jest"))))
	})

	It("rejects a framework that was not detected", func() {
		_, err := filterFrameworks(detected(), []string{"playwright"})
		Expect(err).To(MatchError(ContainSubstring("not detected")))
	})
})
