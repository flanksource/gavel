package fixtures

import (
	"context"
	"os"
	"path/filepath"

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
})
