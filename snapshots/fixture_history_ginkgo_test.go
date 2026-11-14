package snapshots

import (
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture history", func() {
	DescribeTable("identifies historical fixture runs",
		func(args map[string]any, tests []parsers.Test, kind string, frameworks, files []string) {
			snap := testui.Snapshot{Metadata: &testui.SnapshotMetadata{Args: args}, Tests: tests}
			info := runInfo(snap, "run-example.json", "snapshot.json", time.Time{})
			Expect(info.Kind).To(Equal(kind))
			Expect(info.Frameworks).To(Equal(frameworks))
			Expect(info.FixtureFiles).To(Equal(files))
		},
		Entry("recorded arguments before parsing failed", map[string]any{"work_dir": "/repo", "fixture_files": []any{"/repo/hack/purge.fixture.yaml"}}, nil, "fixture", []string{"fixture"}, []string{"hack/purge.fixture.yaml"}),
		Entry("typed invocation arguments", map[string]any{"work_dir": "/repo", "fixture_files": []string{"./hack/purge.fixture.yaml", "/repo/hack/purge.fixture.yaml"}}, nil, "fixture", []string{"fixture"}, []string{"hack/purge.fixture.yaml"}),
		Entry("nested historical tree", nil, []parsers.Test{{Children: []parsers.Test{{Framework: parsers.Framework("fixture"), File: "./purge.fixture.yaml", Passed: true}}}}, "fixture", []string{"fixture"}, []string{"purge.fixture.yaml"}),
		Entry("mixed frameworks", nil, []parsers.Test{{Framework: parsers.Framework("fixture"), File: "purge.fixture.yaml"}, {Framework: parsers.Framework("ginkgo"), Passed: true}}, "test", []string{"fixture", "ginkgo"}, []string{"purge.fixture.yaml"}),
		Entry("multiple fixture files", map[string]any{"fixture_files": []any{"b.fixture.yaml", "a.fixture.yaml"}}, nil, "fixture", []string{"fixture"}, []string{"a.fixture.yaml", "b.fixture.yaml"}),
	)

	It("keeps fixture and lint runs classified as combined", func() {
		info := runInfo(testui.Snapshot{Status: testui.SnapshotStatus{LintRun: true}, Tests: []parsers.Test{{Framework: parsers.Framework("fixture"), Passed: true}}}, "run-example.json", "snapshot.json", time.Time{})
		Expect(info.Kind).To(Equal(RunKindTestLint))
		Expect(info.Frameworks).To(Equal([]string{"fixture"}))
	})
})
