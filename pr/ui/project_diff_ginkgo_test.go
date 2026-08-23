package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/status"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project working-tree diff", func() {
	var originalProjectsPath string

	BeforeEach(func() {
		originalProjectsPath = projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
	})

	AfterEach(func() {
		projectsPath = originalProjectsPath
	})

	It("returns staged, unstaged, and untracked patches for a selected folder", func() {
		dir := createDiffRepository()
		writeDiffFile(dir, "src/staged.go", "package src\n\nvar Staged = true\n")
		runDiffGit(dir, "add", "src/staged.go")
		writeDiffFile(dir, "src/unstaged.go", "package src\n\nvar Unstaged = true\n")
		writeDiffFile(dir, "src/new.go", "package src\n\nvar New = true\n")
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: dir}})).To(Succeed())
		originalGather := gatherProjectStatus
		gatherProjectStatus = func(workDir string, opts status.Options) (*status.Result, error) {
			Expect(opts.NoResults).To(BeTrue())
			return originalGather(workDir, opts)
		}
		DeferCleanup(func() { gatherProjectStatus = originalGather })

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/projects/gavel/diff?path=src", nil)
		request.SetPathValue("name", "gavel")
		(&Server{}).handleProjectDiff(recorder, request)

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var response projectDiffResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response.Path).To(Equal("src"))
		Expect(response.Diff).To(ContainSubstring("var Staged = true"))
		Expect(response.Diff).To(ContainSubstring("var Unstaged = true"))
		Expect(response.Diff).To(ContainSubstring("var New = true"))
		Expect(strings.Index(response.Diff, "src/new.go")).To(BeNumerically("<", strings.Index(response.Diff, "src/staged.go")))
	})

	It("expands a wholly untracked directory into a patch per file", func() {
		dir := createDiffRepository()
		writeDiffFile(dir, "src/devtools/.gitignore", "*.log\n")
		writeDiffFile(dir, "src/devtools/panel.ts", "export const panel = true;\n")
		writeDiffFile(dir, "src/devtools/nested/inspector.ts", "export const inspector = true;\n")
		writeDiffFile(dir, "src/devtools/noise.log", "ignored output\n")
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: dir}})).To(Succeed())

		response := requestProjectDiff("gavel", "src/devtools")

		Expect(response.Path).To(Equal("src/devtools"))
		Expect(response.Diff).To(ContainSubstring("export const panel = true;"))
		Expect(response.Diff).To(ContainSubstring("src/devtools/nested/inspector.ts"))
		Expect(response.Diff).To(ContainSubstring("export const inspector = true;"))
		Expect(response.Diff).NotTo(ContainSubstring("noise.log"))
		Expect(response.Binary).To(BeFalse())
	})

	It("marks a selected binary file without hiding text from folder aggregates", func() {
		dir := createDiffRepository()
		Expect(os.WriteFile(filepath.Join(dir, "src", "image.bin"), []byte{0, 1, 2, 3}, 0o600)).To(Succeed())
		writeDiffFile(dir, "src/text.go", "package src\n\nvar Text = true\n")
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: dir}})).To(Succeed())

		fileResponse := requestProjectDiff("gavel", "src/image.bin")
		Expect(fileResponse.Binary).To(BeTrue())

		folderResponse := requestProjectDiff("gavel", "src")
		Expect(folderResponse.Binary).To(BeFalse())
		Expect(folderResponse.Diff).To(ContainSubstring("Binary files"))
		Expect(folderResponse.Diff).To(ContainSubstring("var Text = true"))
	})

	It("caps large working-tree diffs at a complete line", func() {
		dir := createDiffRepository()
		writeDiffFile(dir, "src/large.txt", strings.Repeat("a changed line with enough content\n", 12_000))
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: dir}})).To(Succeed())

		response := requestProjectDiff("gavel", "src/large.txt")

		Expect(response.Truncated).To(BeTrue())
		Expect(len(response.Diff)).To(BeNumerically("<=", 256*1024))
		Expect(response.Diff).To(HaveSuffix("\n"))
	})

	It("rejects paths outside the current project change tree", func() {
		dir := createDiffRepository()
		Expect(SaveProjects([]Project{{Name: "gavel", Dir: dir}})).To(Succeed())

		for _, path := range []string{"../outside", "/absolute", "missing"} {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/projects/gavel/diff?path="+path, nil)
			request.SetPathValue("name", "gavel")
			(&Server{}).handleProjectDiff(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest), path)
		}
	})
})

func createDiffRepository() string {
	dir := GinkgoT().TempDir()
	runDiffGit(dir, "init", "-q")
	runDiffGit(dir, "config", "user.email", "developer@example.com")
	runDiffGit(dir, "config", "user.name", "Developer")
	writeDiffFile(dir, "src/staged.go", "package src\n\nvar Staged = false\n")
	writeDiffFile(dir, "src/unstaged.go", "package src\n\nvar Unstaged = false\n")
	runDiffGit(dir, "add", "src/staged.go", "src/unstaged.go")
	runDiffGit(dir, "commit", "-qm", "initial")
	return dir
}

func requestProjectDiff(project, path string) projectDiffResponse {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/projects/"+project+"/diff?path="+path, nil)
	request.SetPathValue("name", project)
	(&Server{}).handleProjectDiff(recorder, request)
	Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
	var response projectDiffResponse
	Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
	return response
}

func writeDiffFile(dir, path, content string) {
	abs := filepath.Join(dir, filepath.FromSlash(path))
	Expect(os.MkdirAll(filepath.Dir(abs), 0o700)).To(Succeed())
	Expect(os.WriteFile(abs, []byte(content), 0o600)).To(Succeed())
}

func runDiffGit(dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(output))
}
