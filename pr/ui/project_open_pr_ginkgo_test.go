package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project Open PR queue", func() {
	var (
		server               *Server
		project              Project
		originalProjectsPath string
	)

	BeforeEach(func() {
		originalProjectsPath = projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		project = Project{Name: "gavel", Dir: GinkgoT().TempDir()}
		Expect(SaveProjects([]Project{project})).To(Succeed())

		originalGather := gatherProjectStatus
		gatherProjectStatus = func(string, status.Options) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{{Path: "one.go"}}}, nil
		}
		DeferCleanup(func() {
			gatherProjectStatus = originalGather
			projectsPath = originalProjectsPath
		})
		server = &Server{}
	})

	It("builds a queued Open PR as a validated gavel commit --push command", func() {
		queued, err := server.commitQueueActionArgs(project, projectActionRequest{
			Action: projectActionOpenPR,
			Files:  []string{"one.go"},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(queued).To(Equal(commitQueueRequest{
			action: projectActionOpenPR,
			files:  []string{"one.go"},
			args:   []string{"commit", "--work-dir", project.ResolvedDir(), "--precommit=fail", "--push", "one.go"},
		}))
	})

	It("builds an Open PR without files as a push-only gavel commit that ignores the server's session", func() {
		queued, err := server.commitQueueActionArgs(project, projectActionRequest{Action: projectActionOpenPR})

		Expect(err).NotTo(HaveOccurred())
		Expect(queued).To(Equal(commitQueueRequest{
			action: projectActionOpenPR,
			files:  []string{},
			args:   []string{"commit", "--work-dir", project.ResolvedDir(), "--precommit=fail", "--stage=staged", "--push"},
		}))
	})

	It("still requires files for a plain commit", func() {
		_, err := server.commitQueueActionArgs(project, projectActionRequest{Action: projectActionCommit})
		Expect(err).To(MatchError("commit requires at least one selected file"))
	})

	It("queues a push-only Open PR with an empty file list in its task details", func() {
		pushOnly := "--push"
		runs := newFakeCommitRuns(pushOnly)
		payload, err := json.Marshal(map[string]any{"action": "open-pr", "files": []string{}})
		Expect(err).NotTo(HaveOccurred())
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/projects/gavel/commit-queue", bytes.NewReader(payload))
		request.SetPathValue("name", "gavel")

		server.handleCommitQueue(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusAccepted), recorder.Body.String())
		Eventually(runs.commands).Should(Equal([]string{pushOnly}))
		var run projectCommitRun
		Expect(json.Unmarshal(recorder.Body.Bytes(), &run)).To(Succeed())
		snapshots := clickytask.SnapshotByID(run.RunID)
		Expect(snapshots).To(HaveLen(2))
		Expect(snapshots[1].Name).To(Equal("Open PR"))
		details, err := json.Marshal(snapshots[0].Details)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(details)).To(ContainSubstring(`"action":"open-pr","files":[]`))
		runs.release(pushOnly)
		Eventually(func() string { return clickytask.SnapshotByID(run.RunID)[1].Status }).Should(Equal("success"))
	})

	It("replays a failed push-only Open PR as another push-only request", func() {
		requests, err := retryCommitRequests("run-1", projectCommitGroupDetails{Entries: []projectCommitTaskDetails{
			{TaskID: "task-1", Action: projectActionOpenPR, Files: []string{}},
		}}, map[string]clickytask.Status{"task-1": clickytask.StatusFailed})
		Expect(err).NotTo(HaveOccurred())
		Expect(requests).To(HaveLen(1))

		queued, err := server.commitQueueActionArgs(project, requests[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(queued.args).To(Equal([]string{"commit", "--work-dir", project.ResolvedDir(), "--precommit=fail", "--stage=staged", "--push"}))
	})

	It("rejects missing actions and advanced Open PR options", func() {
		_, err := server.commitQueueActionArgs(project, projectActionRequest{Files: []string{"one.go"}})
		Expect(err).To(MatchError("unknown commit queue action \"\""))

		_, err = server.commitQueueActionArgs(project, projectActionRequest{
			Action:  projectActionOpenPR,
			Options: map[string]any{"files": []any{"one.go"}},
		})
		Expect(err).To(MatchError("advanced options are not supported for open-pr"))
	})

	It("reports the queued action on each entry", func() {
		runs := newFakeCommitRuns("one.go")
		payload, err := json.Marshal(map[string]any{"action": "open-pr", "files": []string{"one.go"}})
		Expect(err).NotTo(HaveOccurred())
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/projects/gavel/commit-queue", bytes.NewReader(payload))
		request.SetPathValue("name", "gavel")

		server.handleCommitQueue(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusAccepted), recorder.Body.String())
		Eventually(runs.commands).Should(Equal([]string{"one.go"}))
		var run projectCommitRun
		Expect(json.Unmarshal(recorder.Body.Bytes(), &run)).To(Succeed())
		snapshots := clickytask.SnapshotByID(run.RunID)
		Expect(snapshots).To(HaveLen(2))
		Expect(snapshots[1].Name).To(Equal("Open PR one.go"))
		Expect(snapshots[0].Details).To(Equal(projectCommitGroupDetails{Entries: []projectCommitTaskDetails{
			{TaskID: snapshots[1].ID, Action: projectActionOpenPR, Files: []string{"one.go"}},
		}}))
		runs.release("one.go")
		Eventually(func() string { return clickytask.SnapshotByID(run.RunID)[1].Status }).Should(Equal("success"))
	})

	It("publishes separate action and commit-queue request contracts", func() {
		components := projectsOpenAPI()["components"].(map[string]any)
		schemas := components["schemas"].(map[string]any)
		actionProperties := schemas["ProjectActionRequest"].(map[string]any)["properties"].(map[string]any)
		queueProperties := schemas["ProjectCommitQueueRequest"].(map[string]any)["properties"].(map[string]any)
		run := schemas["ProjectCommitRun"].(map[string]any)

		Expect(actionProperties["action"].(map[string]any)["enum"]).To(Equal([]string{"lint", "test"}))
		Expect(queueProperties["action"].(map[string]any)["enum"]).To(Equal([]string{"commit", "open-pr"}))
		Expect(run["required"]).To(Equal([]string{"runId"}))
		Expect(schemas).NotTo(HaveKey("CommitQueueEntry"))
	})
})
