package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"

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
		action, files, args, err := server.commitQueueActionArgs(project, projectActionRequest{
			Action: projectActionOpenPR,
			Files:  []string{"one.go"},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(action).To(Equal(projectActionOpenPR))
		Expect(files).To(Equal([]string{"one.go"}))
		Expect(args).To(Equal([]string{"commit", "--work-dir", project.ResolvedDir(), "--precommit=fail", "--push", "one.go"}))
	})

	It("rejects missing actions and advanced Open PR options", func() {
		_, _, _, err := server.commitQueueActionArgs(project, projectActionRequest{Files: []string{"one.go"}})
		Expect(err).To(MatchError("unknown commit queue action \"\""))

		_, _, _, err = server.commitQueueActionArgs(project, projectActionRequest{
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
		entries := server.commitQueueFor("gavel").Entries
		Expect(entries).To(HaveLen(1))
		Expect(entries[0].Action).To(Equal(projectActionOpenPR))
		Expect(entries[0].Files).To(Equal([]string{"one.go"}))
		runs.release("one.go")
		Eventually(func() bool { return server.commitQueueFor("gavel").Running }).Should(BeFalse())
	})

	It("publishes separate action and commit-queue request contracts", func() {
		components := projectsOpenAPI()["components"].(map[string]any)
		schemas := components["schemas"].(map[string]any)
		actionProperties := schemas["ProjectActionRequest"].(map[string]any)["properties"].(map[string]any)
		queueProperties := schemas["ProjectCommitQueueRequest"].(map[string]any)["properties"].(map[string]any)
		entry := schemas["CommitQueueEntry"].(map[string]any)

		Expect(actionProperties["action"].(map[string]any)["enum"]).To(Equal([]string{"lint", "test"}))
		Expect(queueProperties["action"].(map[string]any)["enum"]).To(Equal([]string{"commit", "open-pr"}))
		Expect(entry["required"]).To(ContainElement("action"))
	})
})
