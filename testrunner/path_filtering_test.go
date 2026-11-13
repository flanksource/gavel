package testrunner

import (
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("starting path filtering", func() {
	DescribeTable("discovers go test packages only under the starting paths",
		func(startingPaths, expectedPrefixes []string, unexpectedPrefix string) {
			projectRoot, err := filepath.Abs("..")
			Expect(err).NotTo(HaveOccurred())
			registry := NewRegistry(projectRoot)
			registry.Register(runners.NewGoTest(projectRoot))
			detected, err := registry.DetectAll(projectRoot, startingPaths)
			Expect(err).NotTo(HaveOccurred())
			opts := detected[parsers.GoTest]
			Expect(opts).NotTo(BeNil())
			Expect(opts.Paths).To(Equal(startingPaths))
			runner, _ := registry.Get(parsers.GoTest)
			packages, err := runner.DiscoverPackages(*opts)
			Expect(err).NotTo(HaveOccurred())
			for _, prefix := range expectedPrefixes {
				Expect(packages).To(ContainElement(HavePrefix(prefix)), "packages: %v", packages)
			}
			for _, pkg := range packages {
				Expect(strings.HasPrefix(pkg, unexpectedPrefix)).To(BeFalse(), "packages: %v", packages)
			}
		},
		Entry("parsers directory only", []string{"./testrunner/parsers"}, []string{"./testrunner/parsers"}, "./claudehistory"),
		Entry("claudehistory directory only", []string{"./claudehistory"}, []string{"./claudehistory"}, "./testrunner/parsers"),
		Entry("multiple starting paths", []string{"./testrunner/parsers", "./claudehistory"}, []string{"./testrunner/parsers", "./claudehistory"}, "./todos"),
	)
})
