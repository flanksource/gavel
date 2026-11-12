package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/flanksource/gavel/projects"

	"github.com/flanksource/gavel/commit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project index staging", func() {
	var (
		dir    string
		server *Server
	)

	BeforeEach(func() {
		useProjectsFile(filepath.Join(GinkgoT().TempDir(), "projects.json"))
		dir = createDiffRepository()
		writeDiffFile(dir, "lines.txt", "a\nb\nc\nd\ne\n")
		runDiffGit(dir, "add", "lines.txt")
		runDiffGit(dir, "commit", "-qm", "lines")
		Expect(seedProjects([]projects.Project{{Name: "gavel", Dir: dir}})).To(Succeed())
		server = &Server{}
	})

	AfterEach(func() {
	})

	It("stages and unstages modified, deleted and untracked entries", func() {
		writeDiffFile(dir, "src/unstaged.go", "package src\n\nvar Unstaged = true\n")
		Expect(os.Remove(filepath.Join(dir, "src", "staged.go"))).To(Succeed())
		writeDiffFile(dir, "fresh/new.txt", "new\n")

		stage := postIndex(server, "stage", `{"paths":["src/unstaged.go","src/staged.go","fresh"]}`)
		Expect(stage.Code).To(Equal(http.StatusOK), stage.Body.String())
		Expect(stage.Body.String()).To(MatchJSON(`{"ok":true}`))
		Expect(gitText(dir, "diff", "--cached", "--name-status")).To(Equal("A\tfresh/new.txt\nD\tsrc/staged.go\nM\tsrc/unstaged.go\n"))

		unstage := postIndex(server, "stage", `{"paths":["src/unstaged.go","src/staged.go","fresh/new.txt"],"unstage":true}`)
		Expect(unstage.Code).To(Equal(http.StatusOK), unstage.Body.String())
		Expect(gitText(dir, "diff", "--cached", "--name-status")).To(BeEmpty())
		Expect(gitText(dir, "status", "--porcelain")).To(Equal(" D src/staged.go\n M src/unstaged.go\n?? fresh/\n"))
	})

	It("unstages both halves of a staged rename", func() {
		runDiffGit(dir, "mv", "src/staged.go", "src/renamed.go")

		response := postIndex(server, "stage", `{"paths":["src/renamed.go"],"unstage":true}`)

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(gitText(dir, "diff", "--cached", "--name-status")).To(BeEmpty())
	})

	It("stages only the selected lines and unstages them again", func() {
		writeDiffFile(dir, "lines.txt", "a\nB\nc\nD\ne\nf\n")

		stage := postIndex(server, "stage/lines", `{"path":"lines.txt","lines":[{"side":"old","line":2},{"side":"new","line":2}]}`)

		Expect(stage.Code).To(Equal(http.StatusOK), stage.Body.String())
		Expect(changedLines(gitText(dir, "diff", "--cached", "--", "lines.txt"))).To(Equal([]string{"-b", "+B"}))
		Expect(changedLines(gitText(dir, "diff", "--", "lines.txt"))).To(Equal([]string{"-d", "+D", "+f"}))

		unstage := postIndex(server, "stage/lines", `{"path":"lines.txt","unstage":true,"lines":[{"side":"new","line":2}]}`)

		Expect(unstage.Code).To(Equal(http.StatusOK), unstage.Body.String())
		Expect(changedLines(gitText(dir, "diff", "--cached", "--", "lines.txt"))).To(Equal([]string{"-b"}))
	})

	It("stages selected lines of an untracked file as a partial new file", func() {
		writeDiffFile(dir, "notes.txt", "one\ntwo\nthree\n")

		response := postIndex(server, "stage/lines", `{"path":"notes.txt","lines":[{"side":"new","line":1},{"side":"new","line":2}]}`)

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(gitText(dir, "show", ":notes.txt")).To(Equal("one\ntwo\n"))
		Expect(changedLines(gitText(dir, "diff", "--", "notes.txt"))).To(Equal([]string{"+three"}))
	})

	DescribeTable("rejects requests it cannot apply exactly",
		func(route, body, message string) {
			writeDiffFile(dir, "lines.txt", "a\nB\nc\nd\ne\n")
			Expect(os.WriteFile(filepath.Join(dir, "image.bin"), []byte{0, 1, 2, 3}, 0o600)).To(Succeed())
			before := gitText(dir, "diff", "--cached")

			response := postIndex(server, route, body)

			Expect(response.Code).To(Equal(http.StatusBadRequest), response.Body.String())
			Expect(response.Body.String()).To(ContainSubstring(message))
			Expect(gitText(dir, "diff", "--cached")).To(Equal(before))
			Expect(gitText(dir, "ls-files", "--", "image.bin")).To(BeEmpty())
		},
		Entry("empty line selection", "stage/lines", `{"path":"lines.txt","lines":[]}`, "no changed lines"),
		Entry("context line", "stage/lines", `{"path":"lines.txt","lines":[{"side":"old","line":1}]}`, "old:1"),
		Entry("binary file lines", "stage/lines", `{"path":"image.bin","lines":[{"side":"new","line":1}]}`, "binary"),
		Entry("path outside the project", "stage", `{"paths":["../outside"]}`, "project-relative"),
		Entry("path that is not a change", "stage", `{"paths":["README.md"]}`, "not a current project change"),
		Entry("no paths", "stage", `{"paths":[]}`, "at least one path"),
		Entry("unknown field", "stage", `{"paths":["lines.txt"],"force":true}`, "invalid json"),
		Entry("unstaging an untracked file", "stage", `{"paths":["image.bin"],"unstage":true}`, "untracked"),
		Entry("discarding a path that is not a change", "discard", `{"paths":["src/staged.go"]}`, "not a current project change"),
	)

	It("discards tracked changes to HEAD and removes untracked entries", func() {
		writeDiffFile(dir, "src/staged.go", "package src\n\nvar Staged = true\n")
		runDiffGit(dir, "add", "src/staged.go")
		writeDiffFile(dir, "src/staged.go", "package src\n\nvar Staged = 2\n")
		writeDiffFile(dir, "scratch/a.txt", "a\n")
		writeDiffFile(dir, "loose.txt", "loose\n")
		writeDiffFile(dir, "keep.txt", "keep\n")
		writeDiffFile(dir, "added.txt", "added\n")
		runDiffGit(dir, "add", "added.txt")
		runDiffGit(dir, "mv", "src/unstaged.go", "src/moved.go")

		response := postIndex(server, "discard", `{"paths":["src/staged.go","scratch","loose.txt","added.txt","src/moved.go"]}`)

		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(gitText(dir, "status", "--porcelain")).To(Equal("?? keep.txt\n"))
		Expect(filepath.Join(dir, "scratch")).NotTo(BeAnExistingFile())
		Expect(filepath.Join(dir, "added.txt")).NotTo(BeAnExistingFile())
	})
})

var _ = Describe("committing what is staged through the commit queue", func() {
	It("lands line-level staging of two queued commits without reverting each other", func() {
		useProjectsFile(filepath.Join(GinkgoT().TempDir(), "projects.json"))
		GinkgoT().Setenv("GAVEL_COMMIT_TEST", "1")
		dir := createDiffRepository()
		writeDiffFile(dir, "lines.txt", "a\nb\nc\nd\ne\n")
		runDiffGit(dir, "add", "lines.txt")
		runDiffGit(dir, "commit", "-qm", "lines")
		Expect(seedProjects([]projects.Project{{Name: "gavel", Dir: dir}})).To(Succeed())
		server := &Server{}
		server.SetProjectCommitRunner(func(ctx context.Context, args []string, _ io.Writer) error {
			_, err := commit.Run(ctx, commit.Options{WorkDir: args[2], IndexPatch: args[slices.Index(args, "--index-patch")+1]})
			return err
		})

		writeDiffFile(dir, "lines.txt", "a\nB\nc\nD\ne\nf\n")
		Expect(postIndex(server, "stage/lines", `{"path":"lines.txt","lines":[{"side":"old","line":2},{"side":"new","line":2}]}`).Code).To(Equal(http.StatusOK))
		writeDiffFile(dir, "src/staged.go", "package src\n\nvar Staged = true\n")
		Expect(postIndex(server, "stage", `{"paths":["src/staged.go"]}`).Code).To(Equal(http.StatusOK))

		first := postCommitQueue(server, map[string]any{"action": "commit", "files": []string{"lines.txt"}})
		Expect(first.Code).To(Equal(http.StatusAccepted), first.Body.String())
		second := postCommitQueue(server, map[string]any{"action": "commit", "files": []string{"src/staged.go"}})
		Expect(second.Code).To(Equal(http.StatusAccepted), second.Body.String())
		run := decodeCommitRun(second).RunID
		Eventually(func() []string { return commitTaskStatuses(run) }, "30s").Should(Equal([]string{"success", "success"}))
		Eventually(func() bool {
			queue := gavelCommitQueue(server)
			queue.mu.Lock()
			defer queue.mu.Unlock()
			return queue.current.archived
		}, "30s").Should(BeTrue())

		Expect(gitText(dir, "show", "--name-only", "--pretty=", "HEAD~1")).To(Equal("lines.txt\n"))
		Expect(gitText(dir, "show", "HEAD~1:lines.txt")).To(Equal("a\nB\nc\nd\ne\n"))
		Expect(gitText(dir, "show", "--name-only", "--pretty=", "HEAD")).To(Equal("src/staged.go\n"))
		Expect(gitText(dir, "diff", "--cached")).To(BeEmpty())
		Expect(changedLines(gitText(dir, "diff", "--", "lines.txt"))).To(Equal([]string{"-d", "+D", "+f"}))
	})
})

var _ = Describe("project index OpenAPI", func() {
	It("documents every index route with schemas the document defines", func() {
		document := projectsOpenAPI()
		paths := document["paths"].(map[string]any)
		schemas := document["components"].(map[string]any)["schemas"].(map[string]any)
		for route, method := range map[string]string{
			"/api/projects/{name}/stage": "post", "/api/projects/{name}/stage/lines": "post",
			"/api/projects/{name}/discard": "post", "/api/projects/{name}/open": "post", "/api/editors": "get",
		} {
			Expect(paths).To(HaveKey(route))
			encoded, err := json.Marshal(paths[route].(map[string]any)[method])
			Expect(err).NotTo(HaveOccurred(), route)
			for _, ref := range regexp.MustCompile(`#/components/schemas/(\w+)`).FindAllStringSubmatch(string(encoded), -1) {
				Expect(schemas).To(HaveKey(ref[1]), route)
			}
		}
		diff := schemas["ProjectDiff"].(map[string]any)["properties"].(map[string]any)
		Expect(diff).To(HaveKey("staged"))
		Expect(diff).To(HaveKey("unstaged"))
		Expect(schemas["ProjectIgnoreRequest"].(map[string]any)["properties"]).To(HaveKey("extension"))
	})
})

func postIndex(server *Server, route, body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/projects/gavel/"+route, bytes.NewBufferString(body))
	request.SetPathValue("name", "gavel")
	handlers := map[string]http.HandlerFunc{
		"stage":       server.handleProjectStage,
		"stage/lines": server.handleProjectStageLines,
		"discard":     server.handleProjectDiscard,
	}
	handlers[route](recorder, request)
	return recorder
}

func gitText(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), string(output))
	return string(output)
}

// changedLines are the '+'/'-' body lines of a unified diff.
func changedLines(diff string) []string {
	var changed []string
	for _, line := range bytes.Split([]byte(diff), []byte("\n")) {
		text := string(line)
		if len(text) > 0 && (text[0] == '+' || text[0] == '-') && !bytes.HasPrefix(line, []byte("+++")) && !bytes.HasPrefix(line, []byte("---")) {
			changed = append(changed, text)
		}
	}
	return changed
}
