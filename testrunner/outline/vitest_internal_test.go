package outline

import (
	"os"

	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/testrunner/parsers"
)

var _ = Describe("parseVitestList", func() {
	var entries []*Entry

	BeforeEach(func() {
		data, err := os.ReadFile("testdata/vitest/list-output.json")
		Expect(err).NotTo(HaveOccurred())
		items, err := parseVitestList(data, "/work/ui")
		Expect(err).NotTo(HaveOccurred())
		entries = nil
		for _, item := range items {
			entries = append(entries, vitestEntry(item, "/work"))
		}
	})

	It("splits the suite chain from the title", func() {
		Expect(entries[0].Framework).To(Equal(parsers.Vitest))
		Expect(entries[0].Suite).To(Equal([]string{"lint route grouping"}))
		Expect(entries[0].Name).To(Equal("defaults lint routes to linter-rule-file grouping"))
		Expect(entries[0].File).To(Equal("ui/src/routes.test.ts"))
	})

	It("handles tests without a suite", func() {
		Expect(entries[2].Suite).To(BeNil())
		Expect(entries[2].Name).To(Equal("renders without crashing"))
		Expect(entries[2].Line).To(BeZero())
	})

	It("uses the location line when present", func() {
		Expect(entries[3].Line).To(Equal(12))
		Expect(entries[3].Suite).To(Equal([]string{"located suite"}))
	})
})
