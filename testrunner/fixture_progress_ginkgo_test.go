package testrunner

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons-db/tracing"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture progress adapter", func() {
	It("groups lifecycle steps ahead of the typed execution hierarchy and preserves their states", func() {
		snapshot := fixtures.ExecutionSnapshot{
			Version: fixtures.ExecutionSnapshotVersion,
			State:   fixtures.ExecutionRunning,
			Root: &fixtures.ExecutionNode{
				Key: "fixtures", Name: "Fixtures", Kind: fixtures.ExecutionKindRoot, State: fixtures.ExecutionRunning,
				Children: []*fixtures.ExecutionNode{
					{Key: "setup", Name: "Setup", Kind: fixtures.ExecutionKindSetup, State: fixtures.ExecutionPassed, Duration: time.Second},
					{
						Key: "verification.md", Name: "verification.md", Kind: fixtures.ExecutionKindFile, State: fixtures.ExecutionRunning,
						Children: []*fixtures.ExecutionNode{
							{Key: "tests", Name: "Tests", Kind: fixtures.ExecutionKindTest, State: fixtures.ExecutionRunning, Done: 3, Total: 8},
							{Key: "lint", Name: "Lint", Kind: fixtures.ExecutionKindLint, State: fixtures.ExecutionQueued},
						},
					},
				},
			},
		}

		tests := executionSnapshotToTests(snapshot)

		Expect(tests).To(HaveLen(2))
		Expect(tests[0].Name).To(Equal("Fixture lifecycle"))
		Expect(tests[0].Children).To(HaveLen(1))
		Expect(tests[0].Children[0].Name).To(Equal("Setup"))
		Expect(tests[0].Children[0].Passed).To(BeTrue())
		Expect(tests[0].Children[0].Fixture).To(Equal(&parsers.FixtureExecution{Key: "setup", Kind: "setup", State: "passed"}))
		Expect(tests[1].Name).To(Equal("verification.md"))
		Expect(tests[1].Children).To(HaveLen(2))
		Expect(tests[1].Children[0].Running).To(BeTrue())
		Expect(tests[1].Children[0].Progress).To(Equal(&parsers.TestProgress{Phase: "test", Done: 3, Total: 8}))
		Expect(tests[1].Children[1].Pending).To(BeTrue())
		Expect(tests[1].Children[0].TaskID).To(Equal("tests"))
	})

	It("carries a finished fixture's trace windows into its context, so the UI links them", func() {
		windows := []parsers.TraceResult{{TraceRef: tracing.TraceRef{Kind: "terminal", Stream: "s1", ToSeq: 3}, Count: 3}}
		snapshot := fixtures.ExecutionSnapshot{Root: &fixtures.ExecutionNode{
			Key: "fixtures", Kind: fixtures.ExecutionKindRoot, State: fixtures.ExecutionPassed,
			Children: []*fixtures.ExecutionNode{{
				Key: "hello", Name: "hello", Kind: fixtures.ExecutionKindCommand, State: fixtures.ExecutionPassed, Traces: windows,
			}},
		}}

		Expect(executionSnapshotToTests(snapshot)[0].Context).To(Equal(parsers.FixtureContext{Traces: windows}))
	})

	It("projects a finished step's attempts into its detail, as the verification report does", func() {
		file := filepath.Join(GinkgoT().TempDir(), "retry.fixture.yaml")
		origin := &fixtures.FixtureOrigin{File: file, Line: 7}
		snapshot := fixtures.ExecutionSnapshot{Root: &fixtures.ExecutionNode{
			Key: "fixtures", Kind: fixtures.ExecutionKindRoot, State: fixtures.ExecutionPassed,
			Children: []*fixtures.ExecutionNode{{
				Key: "flaky", Name: "flaky", Kind: fixtures.ExecutionKindCommand, State: fixtures.ExecutionPassed, Origin: origin,
			}},
		}}
		attempts := []fixtures.AttemptReport{
			{Attempt: 1, Status: fixtures.AttemptFailed, Error: "exit 1", DurationMs: 12, Result: &fixtures.FixtureResult{Name: "flaky", Status: task.StatusFAIL}},
			{Attempt: 2, Status: fixtures.AttemptPassed, DurationMs: 9},
		}
		tree := &fixtures.FixtureNode{Children: []*fixtures.FixtureNode{{
			Name: "flaky", Type: fixtures.TestNode, Origin: origin,
			Results: &fixtures.FixtureResult{Name: "flaky", Status: task.StatusPASS, Metadata: map[string]any{
				"step": map[string]any{"id": "flaky", "kind": "exec"}, "attempts": attempts,
			}},
		}}}

		tests := executionSnapshotToTests(snapshot)
		mergeFixtureResultDetails(tests, tree, "")

		raw, err := json.Marshal(tests[0].Detail)
		Expect(err).NotTo(HaveOccurred())
		Expect(raw).To(MatchJSON(`{
			"key": "flaky", "name": "flaky", "kind": "command", "state": "passed",
			"origin": {"file": ` + strconv.Quote(file) + `, "line": 7},
			"step": {"id": "flaky", "kind": "exec"},
			"attempts": [
				{"attempt": 1, "status": "failed", "error": "exit 1", "durationMs": 12,
				 "result": {"name": "flaky", "framework": "fixture", "failed": true}},
				{"attempt": 2, "status": "passed", "durationMs": 9}
			]
		}`))
	})
})
