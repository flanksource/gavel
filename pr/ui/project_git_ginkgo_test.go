package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	clickytask "github.com/flanksource/clicky/task"
	gavelgit "github.com/flanksource/gavel/git"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const pgFeature = "feature/a"

// projectGitFixture is project "acme": a main checkout on main pushed to a bare
// origin, a linked worktree on feature/a with two commits (a.txt +2, b.txt +1)
// and one uncommitted line in b.txt, and a worktree-less branch solo that
// rewrites README.md's only line.
type projectGitFixture struct {
	repo, worktree, bare string
	base                 string
	featureHead          string
	soloHead             string
}

func pgGit(dir string, args ...string) string {
	GinkgoHelper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %s:\n%s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

func pgCommit(dir, name, content, message string) string {
	GinkgoHelper()
	Expect(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)).To(Succeed())
	pgGit(dir, "add", name)
	pgGit(dir, "commit", "-q", "-m", message)
	return pgGit(dir, "rev-parse", "HEAD")
}

// pgCommittedAt is git's own committer date for rev, in UTC.
func pgCommittedAt(dir, rev string) time.Time {
	GinkgoHelper()
	at, err := time.Parse(time.RFC3339, pgGit(dir, "log", "-1", "--format=%cI", rev))
	Expect(err).NotTo(HaveOccurred())
	return at.UTC()
}

func newProjectGitFixture() *projectGitFixture {
	GinkgoHelper()
	GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	GinkgoT().Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	GinkgoT().Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	root, err := filepath.EvalSymlinks(GinkgoT().TempDir())
	Expect(err).NotTo(HaveOccurred())
	f := &projectGitFixture{repo: filepath.Join(root, "acme"), worktree: filepath.Join(root, "acme-feature"), bare: filepath.Join(root, "origin.git")}
	pgGit(root, "init", "-q", "--bare", "-b", "main", f.bare)
	pgGit(root, "init", "-q", "-b", "main", f.repo)
	pgGit(f.repo, "config", "user.email", "test@example.com")
	pgGit(f.repo, "config", "user.name", "test")
	pgGit(f.repo, "remote", "add", "origin", f.bare)
	f.base = pgCommit(f.repo, "README.md", "hello\n", "chore: initial")
	pgGit(f.repo, "push", "-q", "-u", "origin", "main")

	pgGit(f.repo, "worktree", "add", "-q", "-b", pgFeature, f.worktree, f.base)
	pgCommit(f.worktree, "a.txt", "a1\na2\n", "feat: add a")
	f.featureHead = pgCommit(f.worktree, "b.txt", "b\n", "feat: add b")
	Expect(os.WriteFile(filepath.Join(f.worktree, "b.txt"), []byte("b\nmore\n"), 0o644)).To(Succeed())

	pgGit(f.repo, "checkout", "-q", "-b", "solo")
	f.soloHead = pgCommit(f.repo, "README.md", "HELLO\n", "fix: shout hello")
	pgGit(f.repo, "checkout", "-q", "main")

	originalProjectsPath := projectsPath
	projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
	DeferCleanup(func() { projectsPath = originalProjectsPath })
	Expect(SaveProjects([]Project{{Name: "acme", Dir: f.repo}})).To(Succeed())
	return f
}

func serveProjectGit(server *Server, method, target string, body any) *httptest.ResponseRecorder {
	GinkgoHelper()
	var reader *bytes.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		Expect(err).NotTo(HaveOccurred())
		reader = bytes.NewReader(encoded)
	} else {
		reader = bytes.NewReader(nil)
	}
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, httptest.NewRequest(method, target, reader))
	return recorder
}

func decodeProjectGit[T any](recorder *httptest.ResponseRecorder) T {
	GinkgoHelper()
	var out T
	Expect(json.Unmarshal(recorder.Body.Bytes(), &out)).To(Succeed(), recorder.Body.String())
	return out
}

var _ = Describe("project git", func() {
	var f *projectGitFixture
	var server *Server

	BeforeEach(func() {
		f = newProjectGitFixture()
		server = &Server{}
	})

	Describe("GET /api/projects/git-summary", func() {
		It("sums unmerged branch diffs and uncommitted worktree lines, isolating a failing project", func() {
			Expect(SaveProjects([]Project{{Name: "acme", Dir: f.repo}, {Name: "broken", Dir: GinkgoT().TempDir()}})).To(Succeed())

			recorder := serveProjectGit(server, http.MethodGet, "/api/projects/git-summary", nil)

			Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
			summaries := decodeProjectGit[[]projectGitSummary](recorder)
			Expect(summaries).To(HaveLen(2))
			// feature/a: a.txt +2, b.txt +1; solo: README +1 -1; worktree: b.txt +1.
			Expect(summaries[0]).To(Equal(projectGitSummary{Name: "acme", Base: "main", Adds: 5, Dels: 1, Worktrees: 1, Branches: 2}))
			Expect(summaries[1].Name).To(Equal("broken"))
			Expect(summaries[1].Error).To(ContainSubstring("worktree list"))
			Expect(summaries[1]).To(Equal(projectGitSummary{Name: "broken", Error: summaries[1].Error}))
		})
	})

	Describe("GET /api/projects/{name}/git", func() {
		It("reports the base, the primary checkout, each worktree's changes, ages and the unmerged branches", func() {
			touched := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
			Expect(os.Chtimes(filepath.Join(f.worktree, "b.txt"), touched, touched)).To(Succeed())

			recorder := serveProjectGit(server, http.MethodGet, "/api/projects/acme/git", nil)

			Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
			state := decodeProjectGit[projectGitResponse](recorder)
			main := pgGit(f.repo, "rev-parse", "HEAD")
			Expect(state).To(Equal(projectGitResponse{
				Base: "main", CurrentBranch: "main", BaseCheckedOut: true,
				Worktrees: []projectGitWorktree{
					{Worktree: gavelgit.Worktree{Path: f.repo, Branch: "main", Head: main, Primary: true}, LastCommitAt: pgCommittedAt(f.repo, main)},
					{
						Worktree: gavelgit.Worktree{Path: f.worktree, Branch: pgFeature, Head: f.featureHead},
						Changes:  projectGitChanges{Unstaged: 1, Adds: 1}, Ahead: 2,
						LastCommitAt: pgCommittedAt(f.repo, f.featureHead), TouchedAt: &touched,
					},
				},
				Branches: []gavelgit.BranchInfo{
					{Name: pgFeature, Head: f.featureHead, Ahead: 2, Worktree: f.worktree, Diff: gavelgit.DiffStat{Commits: 2, Files: 2, Adds: 3}, LastCommitAt: pgCommittedAt(f.repo, f.featureHead)},
					{Name: "solo", Head: f.soloHead, Ahead: 1, Diff: gavelgit.DiffStat{Commits: 1, Files: 1, Adds: 1, Dels: 1}, LastCommitAt: pgCommittedAt(f.repo, f.soloHead)},
				},
			}))
		})
	})

	Describe("worktree scoping", func() {
		It("scopes status to a linked worktree", func() {
			recorder := serveProjectGit(server, http.MethodGet, "/api/projects/acme/status?worktree="+url.QueryEscape(f.worktree), nil)

			Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
			response := decodeProjectGit[projectStatusResponse](recorder)
			Expect(response.WorkDir).To(Equal(f.worktree))
			Expect(response.Branch).To(Equal(pgFeature))
			Expect(response.Files).To(HaveLen(1))
			Expect(response.Files[0].Path).To(Equal("b.txt"))
		})

		It("scopes the working-tree diff to a linked worktree", func() {
			recorder := serveProjectGit(server, http.MethodGet, "/api/projects/acme/diff?path=b.txt&worktree="+url.QueryEscape(f.worktree), nil)

			Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
			Expect(decodeProjectGit[projectDiffResponse](recorder).Diff).To(ContainSubstring("+more"))
		})

		DescribeTable("rejects a path that is not one of the project's worktrees with 400", func(method, route string) {
			foreign := GinkgoT().TempDir()
			target := strings.ReplaceAll(route, "WT", url.QueryEscape(foreign))
			var body any
			if method == http.MethodPost {
				body = map[string]any{"action": "commit", "files": []string{"b.txt"}, "worktree": foreign}
			}

			recorder := serveProjectGit(server, method, target, body)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
			Expect(decodeProjectGit[map[string]string](recorder)["error"]).To(ContainSubstring("not a worktree of"))
		},
			Entry("status", http.MethodGet, "/api/projects/acme/status?worktree=WT"),
			Entry("diff", http.MethodGet, "/api/projects/acme/diff?path=b.txt&worktree=WT"),
			Entry("commit queue", http.MethodPost, "/api/projects/acme/commit-queue"),
		)

		It("queues a worktree commit in the worktree, on its own queue, and replays it there on retry", func() {
			project, err := GetProject("acme")
			Expect(err).NotTo(HaveOccurred())

			queued, err := server.commitQueueActionArgs(project, projectActionRequest{
				Action: projectActionCommit, Files: []string{"b.txt"}, Worktree: f.worktree,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(queued).To(Equal(commitQueueRequest{
				action: projectActionCommit, files: []string{"b.txt"}, worktree: f.worktree, workDir: f.worktree,
				args: []string{"commit", "--work-dir", f.worktree, "--precommit=fail", "b.txt"},
			}))
			worktreeQueue, primaryQueue := server.projectCommitQueue(project, f.worktree), server.projectCommitQueue(project, f.repo)
			Expect(worktreeQueue).NotTo(BeIdenticalTo(primaryQueue))
			Expect(worktreeQueue.labels(project)).To(Equal(map[string]string{"project": "acme", "action": "commit", "worktree": f.worktree}))
			Expect(primaryQueue.labels(project)).To(Equal(map[string]string{"project": "acme", "action": "commit"}))
			retried, err := retryCommitRequests("run-1", projectCommitGroupDetails{Entries: []projectCommitTaskDetails{
				{TaskID: "t1", Action: projectActionCommit, Files: []string{"b.txt"}, Worktree: f.worktree},
			}}, map[string]clickytask.Status{"t1": clickytask.StatusFailed})
			Expect(err).NotTo(HaveOccurred())
			Expect(retried).To(Equal([]projectActionRequest{{Action: projectActionCommit, Files: []string{"b.txt"}, Worktree: f.worktree}}))
		})
	})

	Describe("branch review", func() {
		It("lists the files a branch changed since it forked from base", func() {
			recorder := serveProjectGit(server, http.MethodGet, "/api/projects/acme/branch/files?branch="+url.QueryEscape(pgFeature), nil)

			Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
			response := decodeProjectGit[todoCommitFilesResponse](recorder)
			Expect(response.Hash).To(Equal(f.featureHead))
			Expect(response.Base).To(Equal(f.base))
			paths := []string{}
			for _, file := range response.Files {
				paths = append(paths, file.Path)
			}
			Expect(paths).To(ConsistOf("a.txt", "b.txt"))
		})

		It("returns one file's branch diff", func() {
			recorder := serveProjectGit(server, http.MethodGet, "/api/projects/acme/branch/diff?branch="+url.QueryEscape(pgFeature)+"&file=a.txt", nil)

			Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
			response := decodeProjectGit[todoCommitDiffResponse](recorder)
			Expect(response.Path).To(Equal("a.txt"))
			Expect(response.Commit).To(Equal(f.featureHead))
			Expect(response.Diff).To(ContainSubstring("+a2"))
			Expect(response.Diff).NotTo(ContainSubstring("b.txt"))
		})

		It("answers 404 for an unknown branch", func() {
			recorder := serveProjectGit(server, http.MethodGet, "/api/projects/acme/branch/files?branch=nope", nil)
			Expect(recorder.Code).To(Equal(http.StatusNotFound), recorder.Body.String())
		})
	})
})
