package fixtures

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture execution progress", func() {
	It("publishes immutable queued, running, and terminal tree snapshots", func() {
		workDir := GinkgoT().TempDir()
		fixtureFile := filepath.Join(workDir, "verification.md")
		root := &FixtureNode{Name: "Fixtures", Type: SectionNode}
		section := &FixtureNode{Name: "Checks", Type: SectionNode}
		testNode := &FixtureNode{
			Name:   "unit tests",
			Type:   TestNode,
			Test:   &FixtureTest{Name: "unit tests", RunnerStep: &RunnerStepSpec{Kind: RunnerKindTest}},
			Origin: &FixtureOrigin{File: fixtureFile, SectionPath: "Checks", Kind: "codeblock", Line: 8},
		}
		lintNode := &FixtureNode{
			Name:   "lint",
			Type:   TestNode,
			Test:   &FixtureTest{Name: "lint", RunnerStep: &RunnerStepSpec{Kind: RunnerKindLint}},
			Origin: &FixtureOrigin{File: fixtureFile, SectionPath: "Checks", Kind: "codeblock", Line: 14},
		}
		root.AddChild(section)
		section.AddChild(testNode)
		section.AddChild(lintNode)

		var snapshots []ExecutionSnapshot
		tracker := newExecutionTracker(root, workDir, nil, func(_ context.Context, snapshot ExecutionSnapshot) error {
			snapshots = append(snapshots, snapshot)
			return nil
		})

		Expect(tracker.Publish(context.Background())).To(Succeed())
		Expect(snapshots).To(HaveLen(1))
		queued := snapshots[0]
		Expect(queued.Version).To(Equal(ExecutionSnapshotVersion))
		Expect(queued.Root.State).To(Equal(ExecutionQueued))
		Expect(queued.Summary).To(Equal(ExecutionSummary{Total: 2, Queued: 2}))
		Expect(queued.Root.Children[0].Children).To(HaveLen(2))
		Expect(queued.Root.Children[0].Children[0].Kind).To(Equal(ExecutionKindTest))
		Expect(queued.Root.Children[0].Children[1].Kind).To(Equal(ExecutionKindLint))
		Expect(queued.Root.Children[0].Children[0].Key).NotTo(Equal(queued.Root.Children[0].Children[1].Key))

		Expect(tracker.Start(context.Background(), testNode)).To(Succeed())
		Expect(snapshots).To(HaveLen(2))
		Expect(snapshots[1].Root.State).To(Equal(ExecutionRunning))
		Expect(snapshots[1].Summary).To(Equal(ExecutionSummary{Total: 2, Running: 1, Queued: 1}))
		Expect(queued.Root.State).To(Equal(ExecutionQueued), "a later transition must not mutate an earlier snapshot")
		Expect(tracker.Update(context.Background(), testNode, 3, 8)).To(Succeed())
		runningTest := snapshots[len(snapshots)-1].Root.Children[0].Children[0]
		Expect(runningTest.Done).To(Equal(3))
		Expect(runningTest.Total).To(Equal(8))

		pass := FixtureResult{Name: testNode.Name, Status: task.StatusPASS}
		Expect(tracker.Complete(context.Background(), testNode, pass)).To(Succeed())
		Expect(snapshots[len(snapshots)-1].Summary).To(Equal(ExecutionSummary{Total: 2, Passed: 1, Queued: 1}))
		Expect(snapshots[len(snapshots)-1].Root.State).To(Equal(ExecutionRunning))

		Expect(tracker.Start(context.Background(), lintNode)).To(Succeed())
		failure := FixtureResult{Name: lintNode.Name, Status: task.StatusFAIL, Error: "lint failed"}
		Expect(tracker.Complete(context.Background(), lintNode, failure)).To(Succeed())
		terminal := snapshots[len(snapshots)-1]
		Expect(terminal.Root.State).To(Equal(ExecutionFailed))
		Expect(terminal.Summary).To(Equal(ExecutionSummary{Total: 2, Passed: 1, Failed: 1}))
		Expect(terminal.Root.Children[0].Children[1].State).To(Equal(ExecutionFailed))
		Expect(terminal.Root.Children[0].Children[1].Error).To(Equal("lint failed"))
	})

	It("keeps prerequisite work as typed children in the same tree", func() {
		root := &FixtureNode{Name: "Fixtures", Type: SectionNode}
		root.AddChild(&FixtureNode{Name: "command", Type: TestNode, Test: &FixtureTest{Name: "command"}})

		tracker := newExecutionTracker(root, "", []ExecutionStep{
			{Key: "setup:fixture.md", Name: "Setup fixture.md", Kind: ExecutionKindSetup},
			{Key: "build", Name: "Build", Kind: ExecutionKindBuild},
			{Key: "daemon", Name: "Daemon", Kind: ExecutionKindDaemon},
		}, nil)
		snapshot := tracker.Snapshot()

		Expect(snapshot.Root.Children).To(HaveLen(4))
		Expect(snapshot.Root.Children[0].Kind).To(Equal(ExecutionKindSetup))
		Expect(snapshot.Root.Children[1].Kind).To(Equal(ExecutionKindBuild))
		Expect(snapshot.Root.Children[2].Kind).To(Equal(ExecutionKindDaemon))
		Expect(snapshot.Root.Children[3].Kind).To(Equal(ExecutionKindCommand))
		Expect(snapshot.Summary).To(Equal(ExecutionSummary{Total: 4, Queued: 4}))
	})

	It("keeps source-derived keys unique when replicated nodes share an origin", func() {
		origin := &FixtureOrigin{File: "verification.md", SectionPath: "Checks", Kind: "command", Line: 8}
		root := &FixtureNode{Name: "Fixtures", Type: SectionNode}
		for _, name := range []string{"first replica", "second replica"} {
			section := &FixtureNode{Name: name, Type: SectionNode}
			section.AddChild(&FixtureNode{
				Name: name, Type: TestNode, Test: &FixtureTest{Name: name}, Origin: origin,
			})
			root.AddChild(section)
		}

		snapshot := newExecutionTracker(root, "", nil, nil).Snapshot()
		first := snapshot.Root.Children[0].Children[0].Key
		second := snapshot.Root.Children[1].Children[0].Key

		Expect(first).NotTo(Equal(second))
	})

	It("carries each step's authored source on its node before the step runs", func() {
		path := filepath.Join(GinkgoT().TempDir(), "purge.fixture.yaml")
		Expect(os.WriteFile(path, []byte("kind: Fixture\nsteps:\n  - id: approval\n    kind: form\n"), 0o600)).To(Succeed())
		root := &FixtureNode{Name: "Fixtures", Type: SectionNode}
		root.AddChild(&FixtureNode{Name: "approval", Type: TestNode, Test: &FixtureTest{Name: "approval"},
			Origin: &FixtureOrigin{File: path, Kind: "step", Line: 3, EndLine: 4}})

		node := newExecutionTracker(root, "", nil, nil).Snapshot().Root.Children[0]

		Expect(node.State).To(Equal(ExecutionQueued))
		Expect(node.Source).NotTo(BeNil())
		Expect(node.Source.Text).To(Equal("  - id: approval\n    kind: form"))
	})

	It("attaches fixture results before publishing their terminal state", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "progress.fixture.md")
		content := "---\nbuild: printf build\n---\n\n# Progress\n\n## completes\n\n```bash\nprintf passed\n```\n"
		Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())

		var snapshots []ExecutionSnapshot
		runner, err := NewRunner(RunnerOptions{
			Paths: []string{path}, WorkDir: workDir,
			ProgressSink: func(_ context.Context, snapshot ExecutionSnapshot) error {
				snapshots = append(snapshots, snapshot)
				return nil
			},
		})
		Expect(err).NotTo(HaveOccurred())

		tree, runErr := runner.Run()

		Expect(runErr).NotTo(HaveOccurred())
		Expect(snapshots).NotTo(BeEmpty())
		Expect(snapshots[0].Summary).To(Equal(ExecutionSummary{Total: 2, Queued: 2}))
		Expect(snapshots[len(snapshots)-1].State).To(Equal(ExecutionPassed))
		Expect(snapshots[len(snapshots)-1].Summary).To(Equal(ExecutionSummary{Total: 2, Passed: 2}))
		var result *FixtureResult
		tree.Walk(func(node *FixtureNode) {
			if node.Test != nil {
				result = node.Results
			}
		})
		Expect(result).NotTo(BeNil())
		Expect(result.Status).To(Equal(task.StatusPASS))
	})

	It("streams a running command's output onto its node and lands its result when it finishes", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "stream.fixture.md")
		content := "# Stream\n\n## prints twice\n\n```bash\nprintf 'first\\n'; sleep 1; printf 'second\\n'\n```\n"
		Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())

		var mu sync.Mutex
		var running []string
		var last ExecutionSnapshot
		runner, err := NewRunner(RunnerOptions{
			Paths: []string{path}, WorkDir: workDir,
			ProgressSink: func(_ context.Context, snapshot ExecutionSnapshot) error {
				mu.Lock()
				defer mu.Unlock()
				walkExecution(snapshot.Root, func(node *ExecutionNode) {
					if node.State == ExecutionRunning && node.Output != "" {
						running = append(running, node.Output)
					}
				})
				last = snapshot
				return nil
			},
		})
		Expect(err).NotTo(HaveOccurred())
		_, runErr := runner.Run()
		Expect(runErr).NotTo(HaveOccurred())

		mu.Lock()
		defer mu.Unlock()
		Expect(running).To(ContainElement("first\n"), "the first line is published while the command still runs")
		var finished *ExecutionNode
		walkExecution(last.Root, func(node *ExecutionNode) {
			if node.Result != nil {
				finished = node
			}
		})
		Expect(finished).NotTo(BeNil())
		Expect(finished.Output).To(BeEmpty(), "a finished node's output is its result's")
		Expect(finished.Result.Stdout).To(Equal("first\nsecond\n"))
	})

	It("coalesces a burst of output into one snapshot and flushes the rest after the interval", func() {
		root := &FixtureNode{Name: "Fixtures", Type: SectionNode}
		node := &FixtureNode{Name: "chatty", Type: TestNode, Test: &FixtureTest{Name: "chatty"}}
		root.AddChild(node)
		var mu sync.Mutex
		var outputs []string
		tracker := newExecutionTracker(root, "", nil, func(_ context.Context, snapshot ExecutionSnapshot) error {
			mu.Lock()
			defer mu.Unlock()
			outputs = append(outputs, snapshot.Root.Children[0].Output)
			return nil
		})
		ctx := context.Background()
		Expect(tracker.Start(ctx, node)).To(Succeed())
		writer := tracker.OutputWriter(ctx, node)
		for _, line := range []string{"a\n", "b\n", "c\n"} {
			_, err := io.WriteString(writer, line)
			Expect(err).NotTo(HaveOccurred())
		}
		read := func() []string {
			mu.Lock()
			defer mu.Unlock()
			return slices.Clone(outputs)
		}
		Expect(read()).To(Equal([]string{"", "a\n"}), "the start, then only the first write of the burst")
		Eventually(read).WithTimeout(2 * time.Second).Should(Equal([]string{"", "a\n", "a\nb\nc\n"}))
		Consistently(read).WithTimeout(2 * outputPublishInterval).Should(HaveLen(3))
	})

	It("publishes a running fixture's live metadata, copied, and drops it once the fixture finishes", func() {
		root := &FixtureNode{Name: "Fixtures", Type: SectionNode}
		node := &FixtureNode{Name: "inspect", Type: TestNode, Test: &FixtureTest{Name: "inspect"}}
		root.AddChild(node)
		var mu sync.Mutex
		var live []map[string]any
		tracker := newExecutionTracker(root, "", nil, func(_ context.Context, snapshot ExecutionSnapshot) error {
			mu.Lock()
			defer mu.Unlock()
			live = append(live, snapshot.Root.Children[0].Metadata)
			return nil
		})
		ctx := context.Background()
		Expect(tracker.Start(ctx, node)).To(Succeed())
		step := &StepContext{evidence: newStepEvidence(withFixtureOutput(ctx, tracker.OutputWriter(ctx, node)), nil)}
		record := map[string]any{"template": "SELECT 1", "connection": map[string]any{"name": "lab", "host": "db.example.com"}}
		Expect(step.SetDetail(QueryMetadataKey, record)).To(Succeed())
		record["sql"] = "SELECT 1"

		mu.Lock()
		first := live[len(live)-1]
		mu.Unlock()
		Expect(first).To(Equal(map[string]any{QueryMetadataKey: map[string]any{"template": "SELECT 1",
			"connection": map[string]any{"name": "lab", "host": "db.example.com"}}}), "the first change publishes at once, as reported then")

		Expect(tracker.Complete(ctx, node, FixtureResult{Name: "inspect", Status: task.StatusPASS})).To(Succeed())
		Expect(tracker.Snapshot().Root.Children[0].Metadata).To(BeNil(), "a finished node's detail is its result's")
		Expect(step.SetDetail(QueryMetadataKey, record)).To(Succeed())
		Expect(tracker.Snapshot().Root.Children[0].Metadata).To(BeNil(), "detail reported after the fixture finished is dropped")
	})

	It("keeps only the tail of a large output, flagged truncated", func() {
		root := &FixtureNode{Name: "Fixtures", Type: SectionNode}
		node := &FixtureNode{Name: "loud", Type: TestNode, Test: &FixtureTest{Name: "loud"}}
		root.AddChild(node)
		tracker := newExecutionTracker(root, "", nil, func(context.Context, ExecutionSnapshot) error { return nil })
		writer := tracker.OutputWriter(context.Background(), node)
		_, err := io.WriteString(writer, strings.Repeat("x", MaxExecutionOutput)+"tail")
		Expect(err).NotTo(HaveOccurred())

		output := tracker.Snapshot().Root.Children[0]
		Expect(output.Output).To(HaveLen(MaxExecutionOutput))
		Expect(output.Output).To(HaveSuffix("xtail"))
		Expect(output.OutputTruncated).To(BeTrue())
	})

	It("carries a finished node's result across the worker's JSON pipe", func() {
		result := &FixtureResult{Name: "inspect", Status: task.StatusWarning, Stdout: "out",
			Metadata: map[string]any{"step": map[string]any{"id": "inspect", "kind": "query"}},
			Children: []*FixtureNode{{Name: "finding", Type: TestNode, Results: &FixtureResult{Name: "finding", Status: task.StatusWarning}}}}
		raw, err := json.Marshal(workerEvent{Version: 1, Kind: "progress", Snapshot: &ExecutionSnapshot{Root: &ExecutionNode{Key: "k", Result: result}}})
		Expect(err).NotTo(HaveOccurred())
		var event workerEvent
		Expect(json.Unmarshal(raw, &event)).To(Succeed())
		Expect(event.Snapshot.Root.Result).To(Equal(result))
	})

	It("discards output written outside a tracked fixture", func() {
		Expect(fixtureOutput(context.Background())).To(BeIdenticalTo(io.Discard))
	})

	It("records the pid of a fixture's command on its progress node", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "pid.fixture.md")
		content := "# Pid\n\n## prints its pid\n\n```bash\nprintf $$\n```\n"
		Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())

		var last ExecutionSnapshot
		runner, err := NewRunner(RunnerOptions{
			Paths: []string{path}, WorkDir: workDir,
			ProgressSink: func(_ context.Context, snapshot ExecutionSnapshot) error {
				last = snapshot
				return nil
			},
		})
		Expect(err).NotTo(HaveOccurred())
		tree, runErr := runner.Run()
		Expect(runErr).NotTo(HaveOccurred())

		var stdout string
		tree.Walk(func(node *FixtureNode) {
			if node.Test != nil {
				stdout = node.Results.Stdout
			}
		})
		var pids []int
		walkExecution(last.Root, func(node *ExecutionNode) {
			pids = append(pids, node.PIDs...)
		})
		Expect(pids).To(HaveLen(1))
		Expect(strconv.Itoa(pids[0])).To(Equal(strings.TrimSpace(stdout)), "the pid the shell printed for itself")
	})
})
