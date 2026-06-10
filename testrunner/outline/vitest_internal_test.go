package outline

import (
	"fmt"
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

var _ = Describe("vitestErrorEntry", func() {
	It("anchors a collection failure on the package's package.json", func() {
		entry := vitestErrorEntry("./site", fmt.Errorf("boom"))
		Expect(entry.Framework).To(Equal(parsers.Vitest))
		Expect(entry.File).To(Equal("site/package.json"))
		Expect(entry.Error).NotTo(BeEmpty())
	})

	It("surfaces the module-resolution cause, stripped of ANSI color", func() {
		raw := fmt.Errorf("vitest list in /abs/site failed: exit status 1\n" +
			"Output:\nvitest.config.ts (1:1) \x1b[33m[UNRESOLVED] \x1b[0mCould not resolve 'vitest/config'")
		Expect(vitestErrorEntry("./site", raw).Error).
			To(Equal("vitest.config.ts (1:1) [UNRESOLVED] Could not resolve 'vitest/config'"))
	})

	It("falls back to the first line when no known cause is present", func() {
		Expect(vitestErrorEntry("ui", fmt.Errorf("exit status 1\nmore detail")).Error).
			To(Equal("exit status 1"))
	})

	It("is excluded from report leaves so it is not counted as a test", func() {
		report := &Report{Entries: []*Entry{
			{Framework: parsers.Vitest, Name: "real test"},
			vitestErrorEntry("./site", fmt.Errorf("boom")),
		}}
		Expect(report.Leaves()).To(HaveLen(1))
		Expect(report.Leaves()[0].Name).To(Equal("real test"))
	})
})
