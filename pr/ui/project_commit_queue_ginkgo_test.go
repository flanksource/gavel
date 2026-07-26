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

// fakeCommitRuns stands in for the `gavel commit` processes the queue spawns.
// Each run is keyed on the single file its group commits, so a spec can observe
// the order commands actually started in and hold one open while it posts the
// next group.
type fakeCommitRuns struct {
	mu      sync.Mutex
	started []string
	gates   map[string]chan struct{}
	failing map[string]bool
}

func newFakeCommitRuns(files ...string) *fakeCommitRuns {
	runs := &fakeCommitRuns{gates: map[string]chan struct{}{}, failing: map[string]bool{}}
	for _, file := range files {
		runs.gates[file] = make(chan struct{})
	}
	original := executeProjectAction
	executeProjectAction = func(_ context.Context, _ string, args []string, output io.Writer, group *clickytask.Group, opts ...clickytask.Option) clickytask.TypedTask[cexec.ExecResult] {
		file := args[len(args)-1]
		return clickytask.StartTask("fake commit "+file, func(ctx commonscontext.Context, _ *clickytask.Task) (cexec.ExecResult, error) {
			runs.mu.Lock()
			runs.started = append(runs.started, file)
			gate, failing := runs.gates[file], runs.failing[file]
			runs.mu.Unlock()
			if _, err := io.WriteString(output, "committing "+file+"\n"); err != nil {
				return cexec.ExecResult{ExitCode: -1}, err
			}
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

var _ = Describe("project commit queue", func() {
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
		// The queue archives its run into the project directory once the last
		// group finishes, so a spec that ends with work still in flight races
		// ginkgo's temp-dir removal. Stop whatever is left and wait for the
		// archive to land before the directory goes away.
		DeferCleanup(func() {
			queue := server.projectCommitQueue("gavel")
			queue.mu.Lock()
			pending := queue.runID != ""
			for _, entry := range queue.entries {
				entry.task.Cancel()
			}
			queue.mu.Unlock()
			if !pending {
				return
			}
			Eventually(func() bool {
				queue.mu.Lock()
				defer queue.mu.Unlock()
				return queue.archived
			}).Should(BeTrue(), "the commit queue should archive its run before the project directory is removed")
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

	cancelEntry := func(id string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodDelete, "/api/projects/gavel/commit-queue/"+id, nil)
		request.SetPathValue("name", "gavel")
		request.SetPathValue("id", id)
		server.handleCommitQueueEntry(recorder, request)
		return recorder
	}

	queueStatus := func() commitQueueStatus { return server.commitQueueFor("gavel") }

	entryStatuses := func() []string {
		var statuses []string
		for _, entry := range queueStatus().Entries {
			statuses = append(statuses, entry.Status)
		}
		return statuses
	}

	It("runs queued groups one at a time in the order they were posted", func() {
		runs := newFakeCommitRuns("one.go", "two.go")

		Expect(queueCommit("one.go").Code).To(Equal(http.StatusAccepted))
		Eventually(runs.commands).Should(Equal([]string{"one.go"}))
		Expect(queueCommit("two.go").Code).To(Equal(http.StatusAccepted))
		Consistently(runs.commands).Should(Equal([]string{"one.go"}))
		Eventually(entryStatuses).Should(Equal([]string{"running", "pending"}))
		Expect(queueStatus().Entries[0].Files).To(Equal([]string{"one.go"}))
		Expect(queueStatus().Entries[0].Output).To(Equal("committing one.go\n"))
		Expect(queueStatus().Running).To(BeTrue())

		runs.release("one.go")
		Eventually(runs.commands).Should(Equal([]string{"one.go", "two.go"}))
		runs.release("two.go")
		Eventually(entryStatuses).Should(Equal([]string{"success", "success"}))
		Expect(queueStatus().Running).To(BeFalse())
	})

	// Commit order is the whole point of building groups by hand: each group is
	// staged against the tree the previous one left behind, so a queue that
	// reorders them writes a history the user did not ask for.
	It("keeps the posted order once several groups are waiting behind the running one", func() {
		posted := []string{"one.go", "two.go", "three.go", "four.go"}
		runs := newFakeCommitRuns(posted...)
		for _, file := range posted {
			Expect(queueCommit(file).Code).To(Equal(http.StatusAccepted))
		}

		for index, file := range posted {
			Eventually(runs.commands).Should(Equal(posted[:index+1]), "group %d should have started next", index+1)
			Consistently(runs.commands).Should(Equal(posted[:index+1]))
			runs.release(file)
		}
		Eventually(entryStatuses).Should(Equal([]string{"success", "success", "success", "success"}))
	})

	// Dropping one group is not a decision about the groups behind it: those
	// files are still selected, still staged in the user's head, and must still
	// be committed in the order they were posted.
	It("keeps the groups behind a dropped one queued", func() {
		posted := []string{"one.go", "two.go", "three.go", "four.go"}
		runs := newFakeCommitRuns(posted...)
		for _, file := range posted {
			Expect(queueCommit(file).Code).To(Equal(http.StatusAccepted))
		}
		Eventually(runs.commands).Should(Equal([]string{"one.go"}))

		Expect(cancelEntry(queueStatus().Entries[1].ID).Code).To(Equal(http.StatusOK))
		runs.release("one.go")
		Eventually(runs.commands).Should(Equal([]string{"one.go", "three.go"}))
		runs.release("three.go")
		Eventually(runs.commands).Should(Equal([]string{"one.go", "three.go", "four.go"}))
		runs.release("four.go")
		Eventually(entryStatuses).Should(Equal([]string{"success", "success", "success"}))
	})

	// A failing commit stops the whole queue, not just the group immediately
	// behind it: the later groups were staged against a tree that never came to
	// be, so committing them would write a history nobody reviewed.
	It("cancels the groups behind a failing commit instead of committing them", func() {
		runs := newFakeCommitRuns("one.go", "two.go", "three.go")
		runs.failOn("one.go")

		for _, file := range []string{"one.go", "two.go", "three.go"} {
			Expect(queueCommit(file).Code).To(Equal(http.StatusAccepted))
		}
		Eventually(runs.commands).Should(Equal([]string{"one.go"}))

		runs.release("one.go")
		Eventually(entryStatuses).Should(Equal([]string{"failed", "canceled", "canceled"}))
		Consistently(runs.commands).Should(Equal([]string{"one.go"}))
		Expect(queueStatus().Entries[0].Error).To(ContainSubstring("commit one.go failed"))
		Expect(queueStatus().Running).To(BeFalse())
	})

	It("drops a queued group so its commit never runs", func() {
		runs := newFakeCommitRuns("one.go", "two.go")

		Expect(queueCommit("one.go").Code).To(Equal(http.StatusAccepted))
		Expect(queueCommit("two.go").Code).To(Equal(http.StatusAccepted))
		Eventually(entryStatuses).Should(Equal([]string{"running", "pending"}))

		Expect(cancelEntry(queueStatus().Entries[1].ID).Code).To(Equal(http.StatusOK))
		Expect(queueStatus().Entries).To(HaveLen(1))

		runs.release("one.go")
		Eventually(entryStatuses).Should(Equal([]string{"success"}))
		Consistently(runs.commands).Should(Equal([]string{"one.go"}))
	})

	// Stopping a commit that is taking too long is not the same as a commit that
	// broke: everything still queued behind it stays queued.
	It("cancels the running commit and lets the groups behind it proceed", func() {
		runs := newFakeCommitRuns("one.go", "two.go", "three.go")

		Expect(queueCommit("one.go").Code).To(Equal(http.StatusAccepted))
		Expect(queueCommit("two.go").Code).To(Equal(http.StatusAccepted))
		Expect(queueCommit("three.go").Code).To(Equal(http.StatusAccepted))
		Eventually(runs.commands).Should(Equal([]string{"one.go"}))

		Expect(cancelEntry(queueStatus().Entries[0].ID).Code).To(Equal(http.StatusOK))
		Eventually(runs.commands).Should(Equal([]string{"one.go", "two.go"}))
		runs.release("two.go")
		Eventually(runs.commands).Should(Equal([]string{"one.go", "two.go", "three.go"}))
		runs.release("three.go")
		Eventually(entryStatuses).Should(Equal([]string{"success", "success"}))
	})

	It("reports unknown commit groups instead of silently succeeding", func() {
		newFakeCommitRuns("one.go")
		Expect(queueCommit("one.go").Code).To(Equal(http.StatusAccepted))

		recorder := cancelEntry("not-a-task")
		Expect(recorder.Code).To(Equal(http.StatusNotFound))
		Expect(recorder.Body.String()).To(ContainSubstring("not-a-task"))
	})

	It("rejects a commit group whose files are not current project changes", func() {
		newFakeCommitRuns()
		recorder := queueCommit("missing.go")
		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		Expect(recorder.Body.String()).To(ContainSubstring("missing.go"))
		Expect(queueStatus().Entries).To(BeEmpty())
	})

	It("keeps running commits posted after the queue drained", func() {
		runs := newFakeCommitRuns("one.go", "two.go")

		Expect(queueCommit("one.go").Code).To(Equal(http.StatusAccepted))
		runs.release("one.go")
		Eventually(entryStatuses).Should(Equal([]string{"success"}))

		Expect(queueCommit("two.go").Code).To(Equal(http.StatusAccepted))
		Eventually(runs.commands).Should(Equal([]string{"one.go", "two.go"}))
		runs.release("two.go")
		Eventually(func() bool { return queueStatus().Running }).Should(BeFalse())
	})

	It("publishes the queue on the project status endpoint under a commit task group", func() {
		runs := newFakeCommitRuns("one.go")
		Expect(queueCommit("one.go").Code).To(Equal(http.StatusAccepted))
		Eventually(runs.commands).Should(Equal([]string{"one.go"}))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/projects/gavel/status", nil)
		request.SetPathValue("name", "gavel")
		server.handleProjectStatus(recorder, request)
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())

		var response projectStatusResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response.CommitQueue.Running).To(BeTrue())
		Expect(response.CommitQueue.Href).To(Equal("/tasks/" + response.CommitQueue.RunID))
		Expect(response.CommitQueue.Entries).To(HaveLen(1))
		Expect(response.CommitQueue.Entries[0].Files).To(Equal([]string{"one.go"}))

		snapshots := clickytask.SnapshotByID(response.CommitQueue.RunID)
		Expect(snapshots).NotTo(BeEmpty())
		Expect(snapshots[0].Type).To(Equal("group"))
		Expect(snapshots[0].Kind).To(Equal("gavel-commit"))
		Expect(snapshots[0].Labels).To(HaveKeyWithValue("project", "gavel"))

		runs.release("one.go")
		Eventually(func() bool { return queueStatus().Running }).Should(BeFalse())
	})
})
