package main

import (
	"encoding/json"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/report"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Standalone lint output", func() {
	DescribeTable("serializes complete lint snapshots for artifact consumers",
		func(format string, summary bool) {
			snap := &testui.Snapshot{
				Metadata: &testui.SnapshotMetadata{Kind: "lint"},
				Status:   testui.SnapshotStatus{LintRun: true},
				Lint:     []*linters.LinterResult{{Linter: "tsc", Success: true}, {Linter: "ruff", Skipped: true}},
			}
			withFormat(format, func() {
				payload, err := json.Marshal(lintRunReturnValue(snap, LintOptions{Summary: summary}))
				Expect(err).NotTo(HaveOccurred())
				var decoded testui.Snapshot
				Expect(json.Unmarshal(payload, &decoded)).To(Succeed())
				Expect(decoded.Metadata.Kind).To(Equal("lint"))
				Expect(decoded.Status.LintRun).To(BeTrue())
				Expect(decoded.Lint).To(Equal(snap.Lint))
				var file report.ResultFile
				Expect(json.Unmarshal(payload, &file)).To(Succeed())
				Expect(file.Tests).To(BeEmpty())
				Expect(file.Lint).To(Equal(snap.Lint))
			})
		},
		Entry("JSON", "json", false),
		Entry("CI formats", "json=results.json,markdown=results.md", false),
		Entry("summary option preserves serialized results", "json", true),
	)
	It("preserves terminal rendering and summary limits", func() {
		snap := &testui.Snapshot{Lint: []*linters.LinterResult{{Linter: "tsc", Success: true}}}
		withFormat("pretty", func() {
			Expect(lintRunReturnValue(snap, LintOptions{})).To(Equal(snap.Lint))
			Expect(lintRunReturnValue(snap, LintOptions{Summary: true, SummaryLimit: 3})).To(Equal(linters.NewSummaryView(snap.Lint, 3)))
		})
	})
	It("labels lint-only snapshot reports with the lint summary", func() {
		snap := testui.Snapshot{
			Status: testui.SnapshotStatus{LintRun: true},
			Lint:   []*linters.LinterResult{{Linter: "tsc", Success: true}},
		}
		Expect(snap.Pretty().String()).To(ContainSubstring("Lint summary"))
		Expect(snap.Pretty().String()).NotTo(ContainSubstring("Test summary"))
	})
})
