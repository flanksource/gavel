package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project ignore", func() {
	var workDir string
	var originalProjectsPath string

	BeforeEach(func() {
		originalProjectsPath = projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		workDir = GinkgoT().TempDir()
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: workDir, Repos: []string{"acme/gavel"}}})).To(Succeed())
	})

	AfterEach(func() {
		projectsPath = originalProjectsPath
	})

	It("adds an untracked file to the repository gitignore", func() {
		Expect(os.WriteFile(filepath.Join(workDir, ".gitignore"), []byte("/existing/\n"), 0o644)).To(Succeed())
		stubProjectIgnoreStatus(workDir, []status.FileStatus{{Path: "generated/output.log", State: status.StateUntracked}})

		response := requestProjectIgnore(`{"path":"generated/output.log","directory":false}`)

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		contents, err := os.ReadFile(filepath.Join(workDir, ".gitignore"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("/existing/\n/generated/output.log\n"))
		var payload projectIgnoreResponse
		Expect(json.Unmarshal(response.Body.Bytes(), &payload)).To(Succeed())
		Expect(payload).To(Equal(projectIgnoreResponse{
			Path: "generated/output.log", Rule: "/generated/output.log", Added: true,
		}))
	})

	It("adds a directory rule when all changed descendants are untracked", func() {
		stubProjectIgnoreStatus(workDir, []status.FileStatus{
			{Path: "generated/output.log", State: status.StateUntracked},
			{Path: "generated/cache/data.json", State: status.StateUntracked},
		})

		response := requestProjectIgnore(`{"path":"generated","directory":true}`)

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		contents, err := os.ReadFile(filepath.Join(workDir, ".gitignore"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("/generated/\n"))
	})

	It("rejects a tracked file because gitignore cannot hide it", func() {
		stubProjectIgnoreStatus(workDir, []status.FileStatus{{Path: "tracked.go", State: status.StateUnstaged}})

		response := requestProjectIgnore(`{"path":"tracked.go","directory":false}`)

		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(response.Body.String()).To(ContainSubstring("tracked"))
		_, err := os.Stat(filepath.Join(workDir, ".gitignore"))
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("rejects paths outside the project", func() {
		response := requestProjectIgnore(`{"path":"../outside","directory":false}`)

		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(response.Body.String()).To(ContainSubstring("project-relative"))
	})
})

func stubProjectIgnoreStatus(workDir string, files []status.FileStatus) {
	originalGather := gatherProjectStatus
	gatherProjectStatus = func(dir string, opts status.Options) (*status.Result, error) {
		Expect(dir).To(Equal(workDir))
		Expect(opts.NoRepomap).To(BeTrue())
		return &status.Result{WorkDir: workDir, Files: files}, nil
	}
	DeferCleanup(func() { gatherProjectStatus = originalGather })
}

func requestProjectIgnore(body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/projects/gavel/ignore", bytes.NewBufferString(body))
	request.SetPathValue("name", "gavel")
	(&Server{}).handleProjectIgnore(recorder, request)
	return recorder
}
