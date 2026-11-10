package testrunner

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons-db/tracing"
	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/kinds/query"
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

		tests := projectTests(snapshot)

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

		Expect(projectTests(snapshot)[0].Context).To(Equal(parsers.FixtureContext{Traces: windows}))
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

		test := projectTests(snapshot)[0]
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

		tests := projectTests(snapshot)
		Expect(mergeFixtureResultDetails(tests, tree, "", nil)).To(Succeed())

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

		tests := projectTests(snapshot)
		Expect(mergeFixtureResultDetails(tests, &fixtures.FixtureNode{Children: []*fixtures.FixtureNode{fileNode}}, "", nil)).To(Succeed())

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

var _ = Describe("live fixture step results", func() {
	// testOf is the projected test named name, anywhere under tests.
	var testOf func(tests []parsers.Test, name string) *parsers.Test
	testOf = func(tests []parsers.Test, name string) *parsers.Test {
		for i := range tests {
			if tests[i].Name == name {
				return &tests[i]
			}
			if found := testOf(tests[i].Children, name); found != nil {
				return found
			}
		}
		return nil
	}

	It("shows a finished step's detail and findings while the next step still runs", func() {
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		dbPath := filepath.Join(dir, "data.db")
		db, err := sql.Open("sqlite", dbPath)
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = db.Close() }()
		_, err = db.Exec("CREATE TABLE activities(guid TEXT); INSERT INTO activities VALUES ('a1'), ('a2')")
		Expect(err).NotTo(HaveOccurred())
		writeTestFile(filepath.Join(dir, ".connections", "lab.yaml"), "apiVersion: mission-control.flanksource.com/v1\nkind: Connection\n"+
			"metadata:\n  name: lab\nspec:\n  type: sqlite\n  url:\n    value: "+dbPath+"\n")
		path := filepath.Join(dir, "live.fixture.yaml")
		writeTestFile(path, `kind: Fixture
steps:
  - id: inspect
    kind: query
    connection: lab
    query: SELECT guid FROM activities ORDER BY guid
    findings: "outputs.rows.map(r, {'severity': 'warning', 'message': r.guid})"
  - id: arrived
    kind: query
    connection: lab
    query: SELECT guid FROM activities
    wait: {until: "rowCount == 3", interval: 50ms, timeout: 20s}
`)
		var mu sync.Mutex
		var live *parsers.Test
		runner, err := fixtures.NewRunner(fixtures.RunnerOptions{
			Paths: []string{path}, WorkDir: dir,
			ProgressSink: func(_ context.Context, snapshot fixtures.ExecutionSnapshot) error {
				mu.Lock()
				defer mu.Unlock()
				tests, err := executionSnapshotToTests(snapshot)
				if err != nil {
					return err
				}
				inspect, arrived := testOf(tests, "inspect"), testOf(tests, "arrived")
				if live != nil || inspect == nil || inspect.Running || inspect.Pending || arrived == nil || !arrived.Running {
					return nil
				}
				live = inspect
				// Only now let the second step's wait succeed.
				_, err = db.Exec("INSERT INTO activities VALUES ('a3')")
				return err
			},
		})
		Expect(err).NotTo(HaveOccurred())
		_, err = runner.Run()
		Expect(err).NotTo(HaveOccurred())

		mu.Lock()
		defer mu.Unlock()
		Expect(live).NotTo(BeNil(), "no snapshot had inspect finished while arrived ran")
		Expect(live.Warned).To(BeTrue())
		var names []string
		for _, child := range live.Children {
			names = append(names, child.Name)
		}
		Expect(names).To(Equal([]string{"inspect: a1", "inspect: a2"}))
		raw, err := json.Marshal(live.Detail)
		Expect(err).NotTo(HaveOccurred())
		var detail map[string]any
		Expect(json.Unmarshal(raw, &detail)).To(Succeed())
		Expect(detail).To(HaveKeyWithValue("step", HaveKeyWithValue("id", "inspect")))
		Expect(detail).To(HaveKeyWithValue("outputs", HaveKeyWithValue("rowCount", BeNumerically("==", 2))))
	})

	It("does not lay a step's start result over its live result twice at run end", func() {
		file := filepath.Join(GinkgoT().TempDir(), "q.fixture.yaml")
		origin := &fixtures.FixtureOrigin{File: file, Line: 3}
		finding := &fixtures.FixtureNode{Name: "inspect: a1", Type: fixtures.TestNode, Results: &fixtures.FixtureResult{Name: "inspect: a1", Status: task.StatusWarning}}
		result := &fixtures.FixtureResult{Name: "inspect", Status: task.StatusWarning, Stdout: "live",
			Metadata: map[string]any{"step": map[string]any{"id": "inspect"}}, Children: []*fixtures.FixtureNode{finding}}
		snapshot := fixtures.ExecutionSnapshot{Root: &fixtures.ExecutionNode{
			Key: "fixtures", Kind: fixtures.ExecutionKindRoot, State: fixtures.ExecutionWarned,
			Children: []*fixtures.ExecutionNode{{
				Key: "inspect", Name: "inspect", Kind: fixtures.ExecutionKindCommand, State: fixtures.ExecutionWarned, Origin: origin, Result: result,
			}},
		}}
		final := *result
		final.Stdout, final.Children = "tree", nil
		tree := &fixtures.FixtureNode{Children: []*fixtures.FixtureNode{{Name: "inspect", Type: fixtures.TestNode, Origin: origin, Results: &final,
			Children: []*fixtures.FixtureNode{finding}}}}

		tests := projectTests(snapshot)
		Expect(tests[0].Stdout).To(Equal("live"))
		Expect(mergeFixtureResultDetails(tests, tree, "", liveResultKeys(snapshot))).To(Succeed())

		Expect(tests[0].Stdout).To(Equal("live"), "the start result was already applied from the node")
		Expect(tests[0].Children).To(HaveLen(1))
	})

	It("shows a running step's streamed output as its stdout", func() {
		snapshot := fixtures.ExecutionSnapshot{Root: &fixtures.ExecutionNode{
			Key: "fixtures", Kind: fixtures.ExecutionKindRoot, State: fixtures.ExecutionRunning,
			Children: []*fixtures.ExecutionNode{{
				Key: "setup", Name: "setup", Kind: fixtures.ExecutionKindCommand, State: fixtures.ExecutionRunning, Output: "batch 1: purging\n",
			}},
		}}
		Expect(projectTests(snapshot)[0].Stdout).To(Equal("batch 1: purging\n"))
	})

	It("carries a running query step's live record into its detail under query, as a finished step's", func() {
		record := map[string]any{
			"template":   "EXEC dbo.purge @Days=$(.params.days)",
			"sql":        "EXEC dbo.purge @Days=@p1",
			"arguments":  []any{float64(30)},
			"connection": map[string]any{"name": "lab", "type": "sql_server", "host": "db.example.com", "port": "1433", "user": "reporter", "database": "app"},
			"messages":   []any{map[string]any{"batch": float64(1), "run": float64(1), "message": "purging"}},
		}
		snapshot := fixtures.ExecutionSnapshot{Root: &fixtures.ExecutionNode{
			Key: "fixtures", Kind: fixtures.ExecutionKindRoot, State: fixtures.ExecutionRunning,
			Children: []*fixtures.ExecutionNode{{
				Key: "purge", Name: "purge", Kind: fixtures.ExecutionKindCommand, State: fixtures.ExecutionRunning,
				Metadata: map[string]any{fixtures.QueryMetadataKey: record},
			}},
		}}
		raw, err := json.Marshal(projectTests(snapshot)[0].Detail)
		Expect(err).NotTo(HaveOccurred())
		var detail map[string]any
		Expect(json.Unmarshal(raw, &detail)).To(Succeed())
		Expect(detail).To(HaveKeyWithValue("query", record))
	})
})

// projectTests is executionSnapshotToTests for a snapshot that must project.
func projectTests(snapshot fixtures.ExecutionSnapshot) []parsers.Test {
	GinkgoHelper()
	tests, err := executionSnapshotToTests(snapshot)
	Expect(err).NotTo(HaveOccurred())
	return tests
}

func writeTestFile(path, body string) {
	GinkgoHelper()
	Expect(os.MkdirAll(filepath.Dir(path), 0o700)).To(Succeed())
	Expect(os.WriteFile(path, []byte(body), 0o600)).To(Succeed())
}

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
		tests := projectTests(snapshot)
		Expect(mergeFixtureResultDetails(tests, tree, "", nil)).To(Succeed())
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
