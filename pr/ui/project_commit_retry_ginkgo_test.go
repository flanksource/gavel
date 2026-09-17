package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"time"

	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/internal/taskhistory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// optionsEchoProvider turns a commit's advanced options into arguments that
// show which options reached the command.
type optionsEchoProvider struct{}

func (optionsEchoProvider) Schema(action string) (ProjectActionSchema, error) {
	return ProjectActionSchema{}, fmt.Errorf("schema for %s is not used by commit retry specs", action)
}

func (optionsEchoProvider) Args(_ string, options map[string]any) ([]string, error) {
	message, ok := options["message"].(string)
	if !ok {
		return nil, fmt.Errorf("message option must be a string, got %T", options["message"])
	}
	files, err := projectActionOptionPaths(options["files"])
	if err != nil {
		return nil, err
	}
	return append([]string{"--message=" + message}, files...), nil
}

func retryCommitQueue(server *Server, project, runID string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/projects/"+project+"/commit-queue/"+runID+"/retry", nil)
	request.SetPathValue("name", project)
	request.SetPathValue("runId", runID)
	server.handleCommitQueueRetry(recorder, request)
	return recorder
}

func commitGroupSnapshot(runID string) clickytask.TaskSnapshot {
	snapshots := clickytask.SnapshotByID(runID)
	Expect(snapshots).NotTo(BeEmpty(), "run %s should be live", runID)
	return snapshots[0]
}

func commitTaskStatuses(runID string) []string {
	var statuses []string
	for _, task := range commitTaskSnapshots(runID) {
		statuses = append(statuses, task.Status)
	}
	return statuses
}

func currentCommitRunID(server *Server) string {
	queue := server.projectCommitQueue("gavel")
	queue.mu.Lock()
	defer queue.mu.Unlock()
	Expect(queue.current).NotTo(BeNil())
	return queue.current.runID
}

func emptyArchive() *supervisorTaskSource {
	return &supervisorTaskSource{
		runs: func(context.Context) ([]taskhistory.RunSummary, error) { return nil, nil },
		snapshot: func(_ context.Context, id string) ([]clickytask.TaskSnapshot, error) {
			return nil, fmt.Errorf("archive has no run %s", id)
		},
	}
}

var _ = Describe("project commit run retry", func() {
	var server *Server

	BeforeEach(func() {
		server = setupCommitQueueServer()
	})

	queued := []string{"one.go", "two.go", "three.go"}

	// failGeneration queues every file in queued, fails the named file after the
	// ones before it succeed, and returns the finished run id.
	failGeneration := func(runs *fakeCommitRuns, failing string, statuses []string) string {
		runs.failOn(failing)
		var runID string
		for _, file := range queued {
			recorder := postCommitQueue(server, map[string]any{"action": "commit", "files": []string{file}})
			Expect(recorder.Code).To(Equal(http.StatusAccepted), recorder.Body.String())
			runID = decodeCommitRun(recorder).RunID
		}
		ran := queued[:slices.Index(queued, failing)+1]
		for _, file := range ran {
			runs.release(file)
		}
		Eventually(func() []string { return commitTaskStatuses(runID) }).Should(Equal(statuses))
		Expect(runs.commands()).To(Equal(ran))
		return runID
	}

	It("advertises stop while commits run and nothing to retry after they succeed", func() {
		runs := newFakeCommitRuns("one.go")
		run := decodeCommitRun(postCommitQueue(server, map[string]any{"action": "commit", "files": []string{"one.go"}}))
		Eventually(runs.commands).Should(Equal([]string{"one.go"}))
		Expect(commitGroupSnapshot(run.RunID).Controls).To(Equal([]clickytask.ControlAction{clickytask.ControlStop}))

		runs.release("one.go")

		Eventually(func() []string { return commitTaskStatuses(run.RunID) }).Should(Equal([]string{"success"}))
		Expect(commitGroupSnapshot(run.RunID).Controls).To(BeNil())
	})

	DescribeTable("re-queues only the failed and canceled commits of a finished run",
		func(failing string, statuses, retried []string) {
			runs := newFakeCommitRuns(queued...)
			oldRunID := failGeneration(runs, failing, statuses)
			ran := runs.commands()
			old := commitGroupSnapshot(oldRunID)
			Expect(old.Controls).To(Equal([]clickytask.ControlAction{clickytask.ControlRetry}))

			runs.clearFailure(failing)
			for _, file := range queued {
				if !slices.Contains(ran, file) {
					runs.release(file)
				}
			}
			Expect(clickytask.ControlRun(GinkgoT().Context(), oldRunID, clickytask.ControlRetry)).To(Succeed())

			newRunID := currentCommitRunID(server)
			Expect(newRunID).NotTo(Equal(oldRunID))
			Eventually(runs.commands).Should(Equal(append(slices.Clone(ran), retried...)))
			Eventually(func() []string { return commitTaskStatuses(newRunID) }).Should(HaveEach("success"))
			retry := commitGroupSnapshot(newRunID)
			Expect(retry.Kind).To(Equal(old.Kind))
			Expect(retry.Labels).To(Equal(old.Labels))
			var files []string
			for _, entry := range retry.Details.(projectCommitGroupDetails).Entries {
				files = append(files, entry.Files...)
			}
			Expect(files).To(Equal(retried))
		},
		Entry("a failed first commit retries it and every canceled commit behind it",
			"one.go", []string{"failed", "canceled", "canceled"}, []string{"one.go", "two.go", "three.go"}),
		Entry("a failed second commit leaves the committed first one alone",
			"two.go", []string{"success", "failed", "canceled"}, []string{"two.go", "three.go"}),
	)

	It("retries over HTTP with a new run id", func() {
		runs := newFakeCommitRuns(queued...)
		oldRunID := failGeneration(runs, "one.go", []string{"failed", "canceled", "canceled"})
		runs.clearFailure("one.go")

		recorder := retryCommitQueue(server, "gavel", oldRunID)

		Expect(recorder.Code).To(Equal(http.StatusAccepted), recorder.Body.String())
		var response map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response).To(Equal(map[string]any{"runId": currentCommitRunID(server)}))
		Expect(response["runId"]).NotTo(Equal(oldRunID))
		Eventually(runs.commands).Should(Equal([]string{"one.go", "one.go", "two.go"}))
	})

	It("rejects a retry through another project's route", func() {
		runs := newFakeCommitRuns(queued...)
		oldRunID := failGeneration(runs, "one.go", []string{"failed", "canceled", "canceled"})
		gavel, err := GetProject("gavel")
		Expect(err).NotTo(HaveOccurred())
		Expect(SaveProjects([]Project{gavel, {Name: "other", Dir: GinkgoT().TempDir()}})).To(Succeed())

		recorder := retryCommitQueue(server, "other", oldRunID)

		Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
		Expect(recorder.Body.String()).To(ContainSubstring(`belongs to project \"gavel\"`))
		Consistently(runs.commands).Should(Equal([]string{"one.go"}))
	})

	It("reports an unknown run as not found", func() {
		server.taskSource = emptyArchive()

		recorder := retryCommitQueue(server, "gavel", "never-queued")

		Expect(recorder.Code).To(Equal(http.StatusNotFound), recorder.Body.String())
		Expect(recorder.Body.String()).To(ContainSubstring(`run \"never-queued\" not found`))
	})

	It("rejects a retry whose commits are still queued from an earlier retry", func() {
		runs := newFakeCommitRuns(queued...)
		oldRunID := failGeneration(runs, "one.go", []string{"failed", "canceled", "canceled"})
		runs.clearFailure("one.go")
		Expect(retryCommitQueue(server, "gavel", oldRunID).Code).To(Equal(http.StatusAccepted))
		Eventually(runs.commands).Should(Equal([]string{"one.go", "one.go", "two.go"}))

		duplicate := retryCommitQueue(server, "gavel", oldRunID)

		Expect(duplicate.Code).To(Equal(http.StatusConflict), duplicate.Body.String())
		Expect(duplicate.Body.String()).To(ContainSubstring("two.go, three.go"))
		Consistently(runs.commands).Should(Equal([]string{"one.go", "one.go", "two.go"}))
	})

	It("rejects the whole retry when one of its files is no longer a project change", func() {
		runs := newFakeCommitRuns(queued...)
		oldRunID := failGeneration(runs, "one.go", []string{"failed", "canceled", "canceled"})
		runs.clearFailure("one.go")
		setCommitQueueProjectFiles("one.go", "three.go")

		recorder := retryCommitQueue(server, "gavel", oldRunID)

		Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
		Expect(recorder.Body.String()).To(ContainSubstring(`\"two.go\" is not a current project change`))
		Consistently(runs.commands).Should(Equal([]string{"one.go"}))
	})

	It("replays the advanced options a commit was queued with", func() {
		server.SetProjectActionOptionsProvider(optionsEchoProvider{})
		runs := newFakeCommitRuns("one.go")
		runs.failOn("one.go")
		options := map[string]any{"files": []any{"one.go"}, "message": "fix: one"}
		recorder := postCommitQueue(server, map[string]any{"action": "commit", "options": options})
		Expect(recorder.Code).To(Equal(http.StatusAccepted), recorder.Body.String())
		oldRunID := decodeCommitRun(recorder).RunID
		runs.release("one.go")
		Eventually(func() []string { return commitTaskStatuses(oldRunID) }).Should(Equal([]string{"failed"}))
		Expect(commitGroupSnapshot(oldRunID).Details).To(Equal(projectCommitGroupDetails{Entries: []projectCommitTaskDetails{{
			TaskID: commitTaskSnapshots(oldRunID)[0].ID, Action: projectActionCommit, Files: []string{"one.go"}, Options: options,
		}}}))
		runs.clearFailure("one.go")

		Expect(clickytask.ControlRun(GinkgoT().Context(), oldRunID, clickytask.ControlRetry)).To(Succeed())

		gavel, err := GetProject("gavel")
		Expect(err).NotTo(HaveOccurred())
		argv := []string{"commit", "--work-dir", gavel.ResolvedDir(), "--precommit=fail", "--message=fix: one", "one.go"}
		Eventually(runs.arguments).Should(Equal([][]string{argv, argv}))
	})
})

var _ = Describe("archived project commit run retry", func() {
	const runID = "archived-commit"
	var server *Server

	BeforeEach(func() {
		server = setupCommitQueueServer()
	})

	// archivedCommitRun records a finished commit generation the way the archive
	// returns it: JSON decoded, so Details is a generic map.
	archivedCommitRun := func(mutate func(group, one, two *clickytask.TaskSnapshot)) []clickytask.TaskSnapshot {
		group := clickytask.TaskSnapshot{
			ID: "Commit gavel", GroupID: runID, Type: "group", Status: string(clickytask.StatusFailed),
			Kind: "gavel-commit", Labels: map[string]string{"project": "gavel", "action": "commit"},
			Controls: []clickytask.ControlAction{clickytask.ControlRetry},
			Details: projectCommitGroupDetails{Entries: []projectCommitTaskDetails{
				{TaskID: "archived-one", Action: projectActionCommit, Files: []string{"one.go"}},
				{TaskID: "archived-two", Action: projectActionCommit, Files: []string{"two.go"}},
			}},
		}
		one := clickytask.TaskSnapshot{ID: "archived-one", GroupID: runID, Type: "task", Status: string(clickytask.StatusFailed)}
		two := clickytask.TaskSnapshot{ID: "archived-two", GroupID: runID, Type: "task", Status: string(clickytask.StatusCancelled)}
		mutate(&group, &one, &two)
		encoded, err := json.Marshal([]clickytask.TaskSnapshot{group, one, two})
		Expect(err).NotTo(HaveOccurred())
		var decoded []clickytask.TaskSnapshot
		Expect(json.Unmarshal(encoded, &decoded)).To(Succeed())
		return decoded
	}

	archiveOf := func(snapshots []clickytask.TaskSnapshot) *supervisorTaskSource {
		run := clickytask.RunMeta{ID: runID, Name: "Commit gavel", Status: snapshots[0].Status, StartedAt: time.Now().Add(-time.Hour).Format(time.RFC3339Nano)}
		return &supervisorTaskSource{
			runs: func(context.Context) ([]taskhistory.RunSummary, error) {
				return []taskhistory.RunSummary{{Run: run, ArchivedAt: time.Now().Add(-time.Hour)}}, nil
			},
			snapshot: func(_ context.Context, id string) ([]clickytask.TaskSnapshot, error) {
				Expect(id).To(Equal(runID))
				return snapshots, nil
			},
			retry: server.retryCommitRunControl,
		}
	}

	unchanged := func(*clickytask.TaskSnapshot, *clickytask.TaskSnapshot, *clickytask.TaskSnapshot) {}

	It("re-queues an evicted run's failed commits through the archived source control", func() {
		runs := newFakeCommitRuns("one.go", "two.go")
		runs.release("one.go")
		runs.release("two.go")
		snapshots := archivedCommitRun(unchanged)
		Expect(snapshots[0].Details).To(BeAssignableToTypeOf(map[string]any{}))
		server.taskSource = archiveOf(snapshots)

		Expect(server.taskSource.Control(GinkgoT().Context(), runID, clickytask.ControlRetry)).To(Succeed())

		Eventually(runs.commands).Should(Equal([]string{"one.go", "two.go"}))
		Expect(currentCommitRunID(server)).NotTo(Equal(runID))
	})

	It("keeps non-retry controls of archived runs not found and requires a retry hook", func() {
		source := archiveOf(archivedCommitRun(unchanged))
		Expect(source.Control(GinkgoT().Context(), runID, clickytask.ControlStop)).To(MatchError(`run "archived-commit" not found`))
		source.retry = nil
		Expect(source.Control(GinkgoT().Context(), runID, clickytask.ControlRetry)).To(MatchError(ContainSubstring("no retry handler")))
	})

	DescribeTable("refuses to retry a recorded run it cannot replay",
		func(mutate func(group, one, two *clickytask.TaskSnapshot), expected string) {
			runs := newFakeCommitRuns("one.go", "two.go")
			server.taskSource = archiveOf(archivedCommitRun(mutate))

			_, err := server.retryCommitRun(GinkgoT().Context(), runID)

			Expect(err).To(MatchError(ContainSubstring(expected)))
			Expect(runs.commands()).To(BeEmpty())
		},
		Entry("another kind of run", func(group, _, _ *clickytask.TaskSnapshot) { group.Kind = "gavel-lint" },
			`run "archived-commit" is a "gavel-lint" run, not a commit queue run`),
		Entry("a run that has not finished", func(group, _, _ *clickytask.TaskSnapshot) { group.Status = string(clickytask.StatusRunning) },
			`run "archived-commit" is still running`),
		Entry("a run without a project label", func(group, _, _ *clickytask.TaskSnapshot) { group.Labels = nil },
			`run "archived-commit" has no project label`),
		Entry("a run without commit details", func(group, _, _ *clickytask.TaskSnapshot) { group.Details = nil },
			`run "archived-commit" has no commit entries`),
		Entry("a commit whose task was not recorded", func(_, _, two *clickytask.TaskSnapshot) { two.ID = "unrelated" },
			`run "archived-commit" has no task snapshot for commit archived-two`),
		Entry("a run whose commits all succeeded", func(_, one, two *clickytask.TaskSnapshot) {
			one.Status = string(clickytask.StatusSuccess)
			two.Status = string(clickytask.StatusSuccess)
		}, `run "archived-commit" has no failed commits to retry`),
	)
})
