package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"

	cexec "github.com/flanksource/clicky/exec"
	clickytask "github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeCommitRuns struct {
	mu      sync.Mutex
	started []string
	gates   map[string]chan struct{}
	failing map[string]bool
	output  map[string]string
}

type fakeCommitController struct {
	task *clickytask.Task
}

func (c *fakeCommitController) Actions() []clickytask.ControlAction {
	if status := c.task.Status(); status == clickytask.StatusPending || status == clickytask.StatusRunning {
		return []clickytask.ControlAction{clickytask.ControlStop}
	}
	return nil
}

func (c *fakeCommitController) Control(_ context.Context, action clickytask.ControlAction) error {
	if action != clickytask.ControlStop {
		return fmt.Errorf("fake commit does not support %q", action)
	}
	c.task.Cancel()
	return nil
}

func newFakeCommitRuns(files ...string) *fakeCommitRuns {
	runs := &fakeCommitRuns{
		gates:   map[string]chan struct{}{},
		failing: map[string]bool{},
		output:  map[string]string{},
	}
	for _, file := range files {
		runs.gates[file] = make(chan struct{})
	}
	original := executeProjectAction
	executeProjectAction = func(_ context.Context, _ string, args []string, _ io.Writer, group *clickytask.Group, opts ...clickytask.Option) clickytask.TypedTask[cexec.ExecResult] {
		file := args[len(args)-1]
		task := clickytask.StartTask("fake commit "+file, func(ctx commonscontext.Context, _ *clickytask.Task) (cexec.ExecResult, error) {
			runs.mu.Lock()
			runs.started = append(runs.started, file)
			runs.output[file] = "committing " + file + "\n"
			gate, failing := runs.gates[file], runs.failing[file]
			runs.mu.Unlock()
			select {
			case <-gate:
			case <-ctx.Done():
				return cexec.ExecResult{ExitCode: -1}, ctx.Err()
			}
			if failing {
				return cexec.ExecResult{ExitCode: 1}, fmt.Errorf("commit %s failed", file)
			}
			return cexec.ExecResult{}, nil
		}, append([]clickytask.Option{clickytask.WithGroup(group)}, opts...)...)
		task.SetOutputProvider(func() clickytask.OutputSnapshot {
			runs.mu.Lock()
			defer runs.mu.Unlock()
			return clickytask.OutputSnapshot{Stdout: runs.output[file]}
		})
		task.SetController(&fakeCommitController{task: task.Task})
		return task
	}
	DeferCleanup(func() { executeProjectAction = original })
	return runs
}

func (r *fakeCommitRuns) failOn(file string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failing[file] = true
}

func (r *fakeCommitRuns) release(file string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	close(r.gates[file])
}

func (r *fakeCommitRuns) commands() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.started...)
}

var _ = Describe("project commit task groups", func() {
	var (
		originalProjectsPath string
		server               *Server
	)

	BeforeEach(func() {
		originalProjectsPath = projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: GinkgoT().TempDir()}})).To(Succeed())

		originalGather := gatherProjectStatus
		gatherProjectStatus = func(string, status.Options) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{
				{Path: "one.go"}, {Path: "two.go"}, {Path: "three.go"}, {Path: "four.go"},
			}}, nil
		}
		DeferCleanup(func() {
			gatherProjectStatus = originalGather
			projectsPath = originalProjectsPath
		})

		server = &Server{}
		DeferCleanup(func() {
			queue := server.projectCommitQueue("gavel")
			queue.mu.Lock()
			generation := queue.current
			if generation != nil {
				generation.group.Cancel()
			}
			queue.mu.Unlock()
			if generation == nil {
				return
			}
			Eventually(func() bool {
				queue.mu.Lock()
				defer queue.mu.Unlock()
				return generation.archived
			}).Should(BeTrue(), "the task generation should archive before the project directory is removed")
		})
	})

	queueCommit := func(files ...string) *httptest.ResponseRecorder {
		payload, err := json.Marshal(map[string]any{"action": "commit", "files": files})
		Expect(err).NotTo(HaveOccurred())
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/projects/gavel/commit-queue", bytes.NewReader(payload))
		request.SetPathValue("name", "gavel")
		server.handleCommitQueue(recorder, request)
		return recorder
	}

	decodeRun := func(recorder *httptest.ResponseRecorder) projectCommitRun {
		var run projectCommitRun
		Expect(json.Unmarshal(recorder.Body.Bytes(), &run)).To(Succeed())
		return run
	}

	taskSnapshots := func(runID string) []clickytask.TaskSnapshot {
		var tasks []clickytask.TaskSnapshot
		for _, snapshot := range clickytask.SnapshotByID(runID) {
			if snapshot.Type == "task" {
				tasks = append(tasks, snapshot)
			}
		}
		return tasks
	}

	It("runs posted commits serially and exposes their metadata through the generic task snapshot", func() {
		runs := newFakeCommitRuns("one.go", "two.go")
		first := queueCommit("one.go")
		Expect(first.Code).To(Equal(http.StatusAccepted), first.Body.String())
		run := decodeRun(first)
		Eventually(runs.commands).Should(Equal([]string{"one.go"}))

		second := queueCommit("two.go")
		Expect(second.Code).To(Equal(http.StatusAccepted), second.Body.String())
		Expect(decodeRun(second).RunID).To(Equal(run.RunID))
		Consistently(runs.commands).Should(Equal([]string{"one.go"}))

		Eventually(func(g Gomega) {
			snapshots := clickytask.SnapshotByID(run.RunID)
			g.Expect(snapshots).To(HaveLen(3))
			g.Expect(snapshots[0].Kind).To(Equal("gavel-commit"))
			g.Expect(snapshots[0].Labels).To(HaveKeyWithValue("project", "gavel"))
			g.Expect(snapshots[0].Details).To(Equal(projectCommitGroupDetails{Entries: []projectCommitTaskDetails{
				{TaskID: snapshots[1].ID, Action: projectActionCommit, Files: []string{"one.go"}},
				{TaskID: snapshots[2].ID, Action: projectActionCommit, Files: []string{"two.go"}},
			}}))
			g.Expect(snapshots[1].Name).To(Equal("Commit one.go"))
			g.Expect(snapshots[1].Description).To(Equal("one.go"))
			g.Expect(snapshots[1].Stdout).To(Equal("committing one.go\n"))
			g.Expect(snapshots[1].Controls).To(Equal([]clickytask.ControlAction{clickytask.ControlStop}))
			g.Expect(snapshots[2].Name).To(Equal("Commit two.go"))
			g.Expect(snapshots[2].Status).To(Equal(string(clickytask.StatusPending)))
		}).Should(Succeed())

		runs.release("one.go")
		Eventually(runs.commands).Should(Equal([]string{"one.go", "two.go"}))
		runs.release("two.go")
		Eventually(func() []string {
			return []string{taskSnapshots(run.RunID)[0].Status, taskSnapshots(run.RunID)[1].Status}
		}).Should(Equal([]string{"success", "success"}))
	})

	It("stops one task without stopping the tasks queued behind it", func() {
		runs := newFakeCommitRuns("one.go", "two.go")
		run := decodeRun(queueCommit("one.go"))
		Expect(queueCommit("two.go").Code).To(Equal(http.StatusAccepted))
		Eventually(runs.commands).Should(Equal([]string{"one.go"}))
		firstTask := taskSnapshots(run.RunID)[0]

		Expect(clickytask.ControlTask(GinkgoT().Context(), run.RunID, firstTask.ID, clickytask.ControlStop)).To(Succeed())

		Eventually(runs.commands).Should(Equal([]string{"one.go", "two.go"}))
		runs.release("two.go")
		Eventually(func() string { return taskSnapshots(run.RunID)[1].Status }).Should(Equal("success"))
	})

	It("stops the whole generation through its advertised group control", func() {
		runs := newFakeCommitRuns("one.go", "two.go")
		run := decodeRun(queueCommit("one.go"))
		Expect(queueCommit("two.go").Code).To(Equal(http.StatusAccepted))
		Eventually(runs.commands).Should(Equal([]string{"one.go"}))

		Expect(clickytask.ControlRun(GinkgoT().Context(), run.RunID, clickytask.ControlStop)).To(Succeed())

		Eventually(func() []string {
			tasks := taskSnapshots(run.RunID)
			return []string{tasks[0].Status, tasks[1].Status}
		}).Should(Equal([]string{"canceled", "canceled"}))
		Consistently(runs.commands).Should(Equal([]string{"one.go"}))
	})

	It("cancels every dependent task after a failed commit", func() {
		runs := newFakeCommitRuns("one.go", "two.go", "three.go")
		runs.failOn("one.go")
		run := decodeRun(queueCommit("one.go"))
		Expect(queueCommit("two.go").Code).To(Equal(http.StatusAccepted))
		Expect(queueCommit("three.go").Code).To(Equal(http.StatusAccepted))
		Eventually(runs.commands).Should(Equal([]string{"one.go"}))

		runs.release("one.go")

		Eventually(func() []string {
			tasks := taskSnapshots(run.RunID)
			return []string{tasks[0].Status, tasks[1].Status, tasks[2].Status}
		}).Should(Equal([]string{"failed", "canceled", "canceled"}))
		Consistently(runs.commands).Should(Equal([]string{"one.go"}))
	})

	It("rejects files already owned by an active task in the project generation", func() {
		newFakeCommitRuns("one.go")
		Expect(queueCommit("one.go").Code).To(Equal(http.StatusAccepted))

		duplicate := queueCommit("one.go")

		Expect(duplicate.Code).To(Equal(http.StatusConflict))
		Expect(duplicate.Body.String()).To(ContainSubstring("one.go"))
	})

	It("rejects files that are not current project changes", func() {
		newFakeCommitRuns()
		recorder := queueCommit("missing.go")
		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		Expect(recorder.Body.String()).To(ContainSubstring("missing.go"))
	})

	It("starts a fresh generation after the previous one drains", func() {
		runs := newFakeCommitRuns("one.go", "two.go")
		first := decodeRun(queueCommit("one.go"))
		runs.release("one.go")
		Eventually(func() string { return taskSnapshots(first.RunID)[0].Status }).Should(Equal("success"))

		second := decodeRun(queueCommit("two.go"))

		Expect(second.RunID).NotTo(Equal(first.RunID))
		Eventually(runs.commands).Should(Equal([]string{"one.go", "two.go"}))
		runs.release("two.go")
	})

	It("returns only the generic task run id from the queue endpoint", func() {
		newFakeCommitRuns("one.go")
		recorder := queueCommit("one.go")
		Expect(recorder.Code).To(Equal(http.StatusAccepted), recorder.Body.String())
		var response map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response).To(HaveLen(1))
		Expect(response).To(HaveKey("runId"))
	})
})
