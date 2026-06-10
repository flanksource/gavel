package outline

import (
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/testrunner/parsers"
)

var _ = Describe("Build", func() {
	It("classifies gotest and ginkgo files and sorts by file", func() {
		report, err := Build(Options{WorkDir: "testdata"})
		Expect(err).NotTo(HaveOccurred())

		files := map[string]parsers.Framework{}
		for _, entry := range report.Entries {
			files[entry.File] = entry.Framework
		}
		Expect(files).To(HaveKeyWithValue("gotest/sample_test.go", parsers.GoTest))
		Expect(files).To(HaveKeyWithValue("ginkgo/sample_ginkgo_test.go", parsers.Ginkgo))
		Expect(report.Entries[0].File).To(Equal("ginkgo/sample_ginkgo_test.go"))
	})

	It("sets static descriptions on every leaf", func() {
		report, err := Build(Options{WorkDir: "testdata"})
		Expect(err).NotTo(HaveOccurred())
		for _, leaf := range report.Leaves() {
			Expect(leaf.Description).NotTo(BeEmpty(), "leaf %s has no description", leaf.Name)
		}
	})

	It("limits results with path filters", func() {
		report, err := Build(Options{WorkDir: "testdata", Paths: []string{"gotest"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Entries).NotTo(BeEmpty())
		for _, entry := range report.Entries {
			Expect(entry.File).To(HavePrefix("gotest/"))
		}
	})

	It("limits results with a framework filter", func() {
		report, err := Build(Options{WorkDir: "testdata", Frameworks: []parsers.Framework{parsers.Ginkgo}})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Entries).NotTo(BeEmpty())
		for _, entry := range report.Entries {
			Expect(entry.Framework).To(Equal(parsers.Ginkgo))
		}
	})

	It("rejects unsupported frameworks", func() {
		_, err := Build(Options{WorkDir: "testdata", Frameworks: []parsers.Framework{parsers.Jest}})
		Expect(err).To(MatchError(ContainSubstring("not supported by outline")))
	})
})

var _ = Describe("staticDescription", func() {
	It("humanizes go test names and lists exercised functions", func() {
		desc := staticDescription(&Entry{
			Framework: parsers.GoTest,
			Name:      "TestBuildLocationMap_SkipsVendor",
			calls:     []string{"Expect", "assert.Equal", "BuildLocationMap", "fmt.Sprintf", "history.Load"},
		})
		Expect(desc).To(Equal("build location map skips vendor — exercises BuildLocationMap, history.Load"))
	})

	It("uses the suite chain verbatim for ginkgo", func() {
		desc := staticDescription(&Entry{
			Framework: parsers.Ginkgo,
			Suite:     []string{"Calculator", "when adding"},
			Name:      "adds two positive numbers",
		})
		Expect(desc).To(Equal("Calculator when adding adds two positive numbers"))
	})
})
