package testrunner

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
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

var _ = Describe("fixture detail in the terminal", func() {
	const stderr = "render failed: CALC_RATE target must be numeric"

	// failedStep merges a failed typed step's result into its progress node, as
	// `gavel fixtures` does before printing its failure section.
	failedStep := func(result *fixtures.FixtureResult) parsers.Test {
		file := filepath.Join(GinkgoT().TempDir(), "render.fixture.yaml")
		origin := &fixtures.FixtureOrigin{File: file, Line: 3}
		snapshot := fixtures.ExecutionSnapshot{Root: &fixtures.ExecutionNode{
			Key: "fixtures", Kind: fixtures.ExecutionKindRoot, State: fixtures.ExecutionFailed,
			Children: []*fixtures.ExecutionNode{{
				Key: "render", Name: "render", Kind: fixtures.ExecutionKindCommand, State: fixtures.ExecutionFailed, Origin: origin,
			}},
		}}
		result.Name, result.Status, result.Stderr = "render", task.StatusFAIL, stderr
		tree := &fixtures.FixtureNode{Children: []*fixtures.FixtureNode{{Name: "render", Type: fixtures.TestNode, Origin: origin, Results: result}}}
		tests := executionSnapshotToTests(snapshot)
		Expect(mergeFixtureResultDetails(tests, tree, "")).To(Succeed())
		return tests[0]
	}

	It("shows a failed step's CEL trace and its stderr once, without the step record as JSON", func() {
		celLines := []string{"stdout.size() > 0", "└─ false"}
		test := failedStep(&fixtures.FixtureResult{Error: "cel failed", CELTrace: strings.Join(celLines, "\n"), Metadata: map[string]any{
			"step": map[string]any{"id": "render", "kind": "exec"}, "outputs": map[string]any{"stdout": ""},
			"attempts": []fixtures.AttemptReport{{Attempt: 1, Status: fixtures.AttemptFailed, Error: "cel failed", DurationMs: 40,
				Result: &fixtures.FixtureResult{Name: "render", Status: task.StatusFAIL, Stderr: stderr}}},
		}})

		rendered := test.Pretty().String()

		Expect(rendered).To(ContainSubstring(celLines[0]))
		Expect(rendered).To(ContainSubstring(celLines[1]))
		Expect(strings.Count(rendered, stderr)).To(Equal(1), rendered)
		Expect(rendered).NotTo(ContainSubstring(`"attempts"`))
		Expect(rendered).NotTo(ContainSubstring(`"outputs"`))
		Expect(rendered).NotTo(ContainSubstring("attempt 1"))
		raw, err := json.Marshal(test.Detail)
		Expect(err).NotTo(HaveOccurred())
		Expect(raw).To(ContainSubstring(`"step":`))
		Expect(raw).To(ContainSubstring(`"outputs":`))
		Expect(raw).To(ContainSubstring(`"attempts":`))
		Expect(raw).To(ContainSubstring(`"cel_trace":`))
	})

	It("carries a failed step's source lines into the detail the UI reads", func() {
		source := parsers.SourceSnippet{File: "render.fixture.yaml", Line: 3, EndLine: 4, Language: "yaml", Text: "  - id: render\n    kind: exec"}
		test := failedStep(&fixtures.FixtureResult{Error: "exit 1", Metadata: map[string]any{"source": source}})

		raw, err := json.Marshal(test.Detail)
		Expect(err).NotTo(HaveOccurred())
		var detail struct {
			Name   string                `json:"name"`
			Source parsers.SourceSnippet `json:"source"`
		}
		Expect(json.Unmarshal(raw, &detail)).To(Succeed())
		Expect(detail.Name).To(Equal("render"))
		Expect(detail.Source).To(Equal(source))
	})

	It("lists each attempt of a retried step on one line, with the first line of its error", func() {
		test := failedStep(&fixtures.FixtureResult{Error: "exit 1", Metadata: map[string]any{
			"attempts": []fixtures.AttemptReport{
				{Attempt: 1, Status: fixtures.AttemptFailed, Error: "connection refused\nretrying", DurationMs: 12},
				{Attempt: 2, Status: fixtures.AttemptFailed, Error: "exit 1", DurationMs: 9},
			},
		}})

		rendered := test.Pretty().String()

		Expect(rendered).To(ContainSubstring("attempt 1 failed (12ms): connection refused"))
		Expect(rendered).NotTo(ContainSubstring("retrying"))
		Expect(rendered).To(ContainSubstring("attempt 2 failed (9ms): exit 1"))
	})

	It("summarises a wait that did not succeed and each armed trace's scalar figures", func() {
		summary := map[string]any{
			"count": 4, "duration_ms": 714.3319999999999, "errors": 0,
			"by_op":   map[string]any{"SELECT": 4},
			"queries": []map[string]any{{"sql": "SELECT 1"}},
		}
		test := failedStep(&fixtures.FixtureResult{Error: "wait timed out",
			Traces:   []parsers.TraceResult{{TraceRef: tracing.TraceRef{Kind: "pg_proxy", Stream: "s1", ToSeq: 4}, Count: 4, Summary: summary}},
			Metadata: map[string]any{"wait": fixtures.WaitReport{Outcome: fixtures.WaitTimeout, DurationMs: 30000, Cause: "service not ready"}},
		})

		rendered := test.Pretty().String()

		Expect(rendered).To(ContainSubstring("wait timeout after 30s: service not ready"))
		Expect(rendered).To(ContainSubstring("pg_proxy: 4 records count=4 duration_ms=714.332 errors=0"))
		Expect(rendered).NotTo(ContainSubstring("SELECT"))
	})
})
