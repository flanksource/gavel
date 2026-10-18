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

	It("keeps the internal progress key out of a node's detail; it is the TaskID", func() {
		started, finished := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), time.Date(2026, 1, 2, 3, 4, 7, 0, time.UTC)
		key := "fixtures/codex.fixture.yaml||file|table:0|row:0|line:0|ordinal:0/codex.fixture.yaml|lifecycle|step|table:0|row:8|line:0|ordinal:7"
		snapshot := fixtures.ExecutionSnapshot{Root: &fixtures.ExecutionNode{
			Key: "fixtures", Kind: fixtures.ExecutionKindRoot, State: fixtures.ExecutionFailed,
			Children: []*fixtures.ExecutionNode{{
				Key: key, Name: "watch", Kind: fixtures.ExecutionKindCommand, State: fixtures.ExecutionFailed,
				Origin:    &fixtures.FixtureOrigin{File: "codex.fixture.yaml", SectionPath: "lifecycle", Kind: "step", RowIndex: 8},
				StartedAt: &started, FinishedAt: &finished, Duration: 2 * time.Second, Error: "watcher exited",
			}},
		}}

		test := executionSnapshotToTests(snapshot)[0]
		raw, err := json.Marshal(test.Detail)
		Expect(err).NotTo(HaveOccurred())

		Expect(test.TaskID).To(Equal(key))
		Expect(raw).To(MatchJSON(`{
			"name": "watch", "kind": "command", "state": "failed",
			"origin": {"file": "codex.fixture.yaml", "section_path": "lifecycle", "kind": "step", "row_index": 8},
			"started_at": "2026-01-02T03:04:05Z", "finished_at": "2026-01-02T03:04:07Z",
			"duration": 2000000000, "error": "watcher exited"
		}`))
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
		Expect(mergeFixtureResultDetails(tests, tree, "")).To(Succeed())

		raw, err := json.Marshal(tests[0].Detail)
		Expect(err).NotTo(HaveOccurred())
		Expect(raw).To(MatchJSON(`{
			"name": "flaky", "kind": "command", "state": "passed",
			"origin": {"file": ` + strconv.Quote(file) + `, "line": 7},
			"step": {"id": "flaky", "kind": "exec"},
			"attempts": [
				{"attempt": 1, "status": "failed", "error": "exit 1", "durationMs": 12,
				 "result": {"name": "flaky", "framework": "fixture", "failed": true}},
				{"attempt": 2, "status": "passed", "durationMs": 9}
			]
		}`))
	})

	It("lays a background step's teardown verdict, with its watcher events, over its start result", func() {
		file := filepath.Join(GinkgoT().TempDir(), "codex.fixture.yaml")
		stepOrigin := &fixtures.FixtureOrigin{File: file, SectionPath: "lifecycle", Kind: "step", RowIndex: 8}
		snapshot := fixtures.ExecutionSnapshot{Root: &fixtures.ExecutionNode{
			Key: "fixtures", Kind: fixtures.ExecutionKindRoot, State: fixtures.ExecutionPassed,
			Children: []*fixtures.ExecutionNode{{
				Key: "file", Name: "codex.fixture.yaml", Kind: fixtures.ExecutionKindFile, State: fixtures.ExecutionPassed,
				Origin: &fixtures.FixtureOrigin{File: file},
				Children: []*fixtures.ExecutionNode{{
					Key: "watch-key", Name: "watch", Kind: fixtures.ExecutionKindCommand, State: fixtures.ExecutionPassed, Origin: stepOrigin,
				}},
			}},
		}}
		step := map[string]any{"id": "watch", "kind": "playwright"}
		started := &fixtures.FixtureResult{Name: "watch", Status: task.StatusPASS, Metadata: map[string]any{
			"step": step, "background": true, "outputs": map[string]any{}, "exports": map[string]any{},
		}}
		verdict := fixtures.FixtureResult{Name: "watch", Status: task.StatusPASS, Stdout: "watcher stdout", Metadata: map[string]any{
			"step": step, "background": true, "outputs": map[string]any{}, "exports": map[string]any{}, "teardown": true,
			"events": []map[string]any{{"type": "console", "text": "boom"}},
		}}
		fileNode := &fixtures.FixtureNode{Name: "codex.fixture.yaml", Type: fixtures.FileNode, Origin: &fixtures.FixtureOrigin{File: file}}
		fileNode.AddChild(&fixtures.FixtureNode{Name: "watch", Type: fixtures.TestNode, Origin: stepOrigin, Results: started})
		fileNode.AddChild(&fixtures.FixtureNode{Name: verdict.Name, Type: fixtures.TestNode, Results: &verdict})

		tests := executionSnapshotToTests(snapshot)
		Expect(mergeFixtureResultDetails(tests, &fixtures.FixtureNode{Children: []*fixtures.FixtureNode{fileNode}}, "")).To(Succeed())

		watch := tests[0].Children[0]
		raw, err := json.Marshal(watch.Detail)
		Expect(err).NotTo(HaveOccurred())
		Expect(watch.Stdout).To(Equal("watcher stdout"))
		Expect(raw).To(MatchJSON(`{
			"name": "watch", "kind": "command", "state": "passed",
			"origin": {"file": ` + strconv.Quote(file) + `, "section_path": "lifecycle", "kind": "step", "row_index": 8},
			"step": {"id": "watch", "kind": "playwright"}, "background": true, "outputs": {}, "exports": {},
			"events": [{"type": "console", "text": "boom"}]
		}`))
	})
})
