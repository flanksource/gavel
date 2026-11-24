package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"

	"github.com/flanksource/gavel/projects"

	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// openPRFlagsProvider renders open-pr options as flags in a fixed order; the
// real schema-driven provider lives in cmd/gavel and is specced there.
type openPRFlagsProvider struct{}

func (openPRFlagsProvider) Schema(action string) (ProjectActionSchema, error) {
	return ProjectActionSchema{}, fmt.Errorf("schema for %s is not used by Open PR queue specs", action)
}

func (openPRFlagsProvider) Args(action string, options map[string]any) ([]string, error) {
	if action != string(projectActionOpenPR) {
		return nil, fmt.Errorf("unexpected %s options", action)
	}
	var args []string
	for _, name := range []string{"push", "target", "auto-merge", "merge-type"} {
		if value, ok := options[name]; ok {
			args = append(args, fmt.Sprintf("--%s=%v", name, value))
		}
	}
	for name := range options {
		if !slices.Contains([]string{"push", "target", "auto-merge", "merge-type"}, name) {
			return nil, fmt.Errorf("unsupported open-pr option %q", name)
		}
	}
	return args, nil
}

var _ = Describe("project Open PR queue", func() {
	var (
		server  *Server
		project projects.Project
	)

	BeforeEach(func() {
		useProjectsFile(filepath.Join(GinkgoT().TempDir(), "projects.json"))
		project = projects.Project{Name: "gavel", Dir: GinkgoT().TempDir()}
		initStagedRepo(project.Dir, "one.go")
		Expect(seedProjects([]projects.Project{project})).To(Succeed())

		originalGather := gatherProjectStatus
		gatherProjectStatus = func(string, status.Options) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{{Path: "one.go", State: status.StateStaged}}}, nil
		}
		DeferCleanup(func() {
			gatherProjectStatus = originalGather
		})
		server = &Server{}
	})

	It("builds a queued Open PR as a validated gavel commit --push command", func() {
		queued, err := server.commitQueueActionArgs(project, projectActionRequest{
			Action: projectActionOpenPR,
			Files:  []string{"one.go"},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(queued.indexPatch).To(HaveSuffix(".patch"))
		Expect(queued).To(Equal(commitQueueRequest{
			action:     projectActionOpenPR,
			files:      []string{"one.go"},
			args:       []string{"commit", "--work-dir", project.ResolvedDir(), "--index-patch", queued.indexPatch, "--push"},
			workDir:    project.ResolvedDir(),
			indexPatch: queued.indexPatch,
			snapshot:   []string{"one.go"},
		}))
	})

	It("builds an Open PR without files as a push-only gavel commit that ignores the server's session", func() {
		queued, err := server.commitQueueActionArgs(project, projectActionRequest{Action: projectActionOpenPR})

		Expect(err).NotTo(HaveOccurred())
		Expect(queued).To(Equal(commitQueueRequest{
			action:  projectActionOpenPR,
			files:   []string{},
			args:    []string{"commit", "--work-dir", project.ResolvedDir(), "--stage=staged", "--push"},
			workDir: project.ResolvedDir(),
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
		Expect(queued.args).To(Equal([]string{"commit", "--work-dir", project.ResolvedDir(), "--stage=staged", "--push"}))
	})

	It("rejects a missing action", func() {
		_, err := server.commitQueueActionArgs(project, projectActionRequest{Files: []string{"one.go"}})
		Expect(err).To(MatchError("unknown commit queue action \"\""))
	})

	It("commits the files with default options and pushes with the advanced Open PR flags", func() {
		server.SetProjectActionOptionsProvider(openPRFlagsProvider{})
		options := map[string]any{"push": "branch", "auto-merge": true}
		queued, err := server.commitQueueActionArgs(project, projectActionRequest{Action: projectActionOpenPR, Files: []string{"one.go"}, Options: options})
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{queued.args, queued.options}).To(Equal([]any{
			[]string{"commit", "--work-dir", project.ResolvedDir(), "--index-patch", queued.indexPatch, "--push=branch", "--auto-merge=true"},
			options,
		}))
	})

	It("adds a bare --push to advanced flags that leave the push mode unset, also when pushing only", func() {
		server.SetProjectActionOptionsProvider(openPRFlagsProvider{})
		queued, err := server.commitQueueActionArgs(project, projectActionRequest{Action: projectActionOpenPR, Options: map[string]any{"target": "zim/main"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(queued.args).To(Equal([]string{"commit", "--work-dir", project.ResolvedDir(), "--stage=staged", "--push", "--target=zim/main"}))
	})

	It("refuses invalid advanced Open PR options before queueing", func() {
		server.SetProjectActionOptionsProvider(openPRFlagsProvider{})
		_, err := server.commitQueueActionArgs(project, projectActionRequest{Action: projectActionOpenPR, Files: []string{"one.go"}, Options: map[string]any{"message": "x"}})
		Expect(err).To(MatchError(`unsupported open-pr option "message"`))
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
		Expect(withoutIndexPatches(snapshots[0].Details)).To(Equal(projectCommitGroupDetails{Entries: []projectCommitTaskDetails{
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
