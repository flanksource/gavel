package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo commit endpoints on a run worktree branch", func() {
	var repo, worktree, setup, head string
	gitIn := func(dir string, args ...string) string {
		GinkgoHelper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "git %v: %s", args, out)
		return strings.TrimSpace(string(out))
	}
	commitIn := func(dir, file, content, message string) string {
		GinkgoHelper()
		Expect(os.MkdirAll(filepath.Join(dir, filepath.Dir(file)), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, file), []byte(content), 0o644)).To(Succeed())
		gitIn(dir, "add", file)
		gitIn(dir, "commit", "-q", "-m", message)
		return gitIn(dir, "rev-parse", "HEAD")
	}
	get := func(handler http.HandlerFunc, path string, query url.Values) *httptest.ResponseRecorder {
		GinkgoHelper()
		rec := httptest.NewRecorder()
		handler(rec, httptest.NewRequest(http.MethodGet, path+"?"+query.Encode(), nil))
		return rec
	}

	BeforeEach(func() {
		repo = GinkgoT().TempDir()
		// main is the PR base the git state tracker compares branches to.
		gitIn(repo, "init", "-q", "-b", "main")
		gitIn(repo, "config", "user.email", "test@example.com")
		gitIn(repo, "config", "user.name", "Test User")
		gitIn(repo, "config", "commit.gpgsign", "false")
		commitIn(repo, "README.md", "readme\n", "chore: initial")

		worktree = filepath.Join(GinkgoT().TempDir(), "run-x")
		gitIn(repo, "worktree", "add", "-q", "-b", "run/x", worktree)
		setup = commitIn(worktree, "wip.txt", "wip\n", "chore(setup): snapshot uncommitted changes")
		commitIn(worktree, "pkg/a.go", "package pkg\n", "feat: a")
		head = commitIn(worktree, "docs/b.md", "# b\n", "docs: b")
	})

	It("serves setup..head files and diff from the main checkout", func() {
		tracked := newTrackedGitServer()
		s := tracked.server
		s.ghOpts = github.Options{WorkDir: repo}
		query := url.Values{"dir": {repo}, "base": {setup}, "hash": {head}}

		filesRec := get(s.handleTodoCommitFiles, "/api/todos/commits/files", query)
		Expect(filesRec.Code).To(Equal(http.StatusOK), filesRec.Body.String())
		var files todoCommitFilesResponse
		Expect(json.Unmarshal(filesRec.Body.Bytes(), &files)).To(Succeed())
		var paths []string
		for _, f := range files.Files {
			paths = append(paths, f.Path)
		}
		Expect(paths).To(ConsistOf("pkg/a.go", "docs/b.md"))
		Expect(files.Hash).To(Equal(head))
		Expect(files.Base).To(Equal(setup))
		var cached int64
		Expect(tracked.db.Raw(`SELECT count(*) FROM git_range_stats WHERE base_sha = ? AND head_sha = ? AND file_list IS NOT NULL`,
			setup, head).Scan(&cached).Error).To(Succeed())
		Expect(cached).To(Equal(int64(1)), "the range's file list is stored on first read")

		originalPath := os.Getenv("PATH")
		withoutGit()
		again := get(s.handleTodoCommitFiles, "/api/todos/commits/files", query)
		Expect(again.Code).To(Equal(http.StatusOK), again.Body.String())
		Expect(again.Body.String()).To(Equal(filesRec.Body.String()), "a stored range lists its files without git")
		GinkgoT().Setenv("PATH", originalPath)

		diffRec := get(s.handleTodoCommitDiff, "/api/todos/commits/diff", query)
		Expect(diffRec.Code).To(Equal(http.StatusOK), diffRec.Body.String())
		var diff todoCommitDiffResponse
		Expect(json.Unmarshal(diffRec.Body.Bytes(), &diff)).To(Succeed())
		Expect(diff.Diff).To(ContainSubstring("diff --git a/pkg/a.go b/pkg/a.go"))
		Expect(diff.Diff).To(ContainSubstring("diff --git a/docs/b.md b/docs/b.md"))
		Expect(diff.Diff).NotTo(ContainSubstring("wip.txt"), "the setup snapshot is the range base, not part of it")
		Expect(diff.Commit).To(Equal(head))
	})

	It("answers 410 once the run branch's commits are gone from the repository", func() {
		gitIn(repo, "worktree", "remove", "--force", worktree)
		gitIn(repo, "branch", "-D", "run/x")
		gitIn(repo, "reflog", "expire", "--expire=now", "--all")
		gitIn(repo, "gc", "-q", "--prune=now")

		s := &Server{ghOpts: github.Options{WorkDir: repo}}
		rec := get(s.handleTodoCommitDiff, "/api/todos/commits/diff", url.Values{"dir": {repo}, "base": {setup}, "hash": {head}})
		Expect(rec.Code).To(Equal(http.StatusGone), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring(setup))
	})
})
