package ui

import (
	"context"
	"net/http"
	"strings"

	commitpkg "github.com/flanksource/gavel/commit"
	prcreate "github.com/flanksource/gavel/pr/create"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project branch landing", func() {
	var f *projectGitFixture
	var server *Server
	var created []model.CreatePRInput

	BeforeEach(func() {
		f = newProjectGitFixture()
		created = nil
		server = &Server{prDeps: &prcreate.Deps{Preflight: func(context.Context, provider.Options) error { return nil },
			GenerateContent: func(_ context.Context, in commitpkg.PRContentInput) (commitpkg.PRContent, error) {
				return commitpkg.PRContent{Title: "fix: shout the greeting", Body: "Summarises " + in.Commits[0].Message, Branch: "fix/shout"}, nil
			},
			CreatePR: func(_ provider.Options, in model.CreatePRInput) (*model.CreatePRResult, error) {
				created = append(created, in)
				return &model.CreatePRResult{Number: 7, URL: "https://github.com/acme/acme/pull/7", Base: in.Base}, nil
			},
		}}
	})

	merge := func(body map[string]any) (int, map[string]any) {
		GinkgoHelper()
		recorder := serveProjectGit(server, http.MethodPost, "/api/projects/acme/branch/merge", body)
		return recorder.Code, decodeProjectGit[map[string]any](recorder)
	}

	It("refuses to merge a branch whose worktree has uncommitted changes", func() {
		code, response := merge(map[string]any{"branch": pgFeature, "mode": "incremental"})

		Expect(code).To(Equal(http.StatusConflict))
		Expect(response["error"]).To(ContainSubstring("uncommitted changes"))
		Expect(pgGit(f.repo, "rev-parse", "HEAD")).To(Equal(f.base))
	})

	It("merges a worktree branch commit by commit, then removes its worktree and branch", func() {
		pgGit(f.worktree, "checkout", "--", "b.txt")

		code, response := merge(map[string]any{"branch": pgFeature, "mode": "incremental"})

		Expect(code).To(Equal(http.StatusOK), "%v", response)
		head := pgGit(f.repo, "rev-parse", "HEAD")
		Expect(response).To(Equal(map[string]any{
			"targetBranch": "main", "landedSha": head, "mode": "incremental", "commits": float64(2),
			"worktreeRemoved": true, "branchDeleted": true,
		}))
		Expect(pgGit(f.repo, "log", "--format=%s", f.base+"..HEAD")).To(Equal("feat: add b\nfeat: add a"))
		Expect(f.worktree).NotTo(BeAnExistingFile())
		Expect(pgGit(f.repo, "branch", "--list", pgFeature)).To(BeEmpty())
	})

	It("squashes a branch with the given message and deletes it", func() {
		code, response := merge(map[string]any{"branch": "solo", "mode": "squash", "message": "fix: louder hello"})

		Expect(code).To(Equal(http.StatusOK), "%v", response)
		Expect(response).To(Equal(map[string]any{
			"targetBranch": "main", "landedSha": pgGit(f.repo, "rev-parse", "HEAD"), "mode": "squash", "commits": float64(1),
			"worktreeRemoved": false, "branchDeleted": true,
		}))
		Expect(pgGit(f.repo, "log", "-1", "--format=%B")).To(Equal("fix: louder hello"))
		Expect(pgGit(f.repo, "rev-parse", "HEAD~1")).To(Equal(f.base))
		Expect(pgGit(f.repo, "branch", "--list", "solo")).To(BeEmpty())
	})

	It("generates the squash message from the branch's commits when none is given", func() {
		code, response := merge(map[string]any{"branch": "solo", "mode": "squash"})

		Expect(code).To(Equal(http.StatusOK), "%v", response)
		Expect(pgGit(f.repo, "log", "-1", "--format=%B")).To(Equal("fix: shout the greeting\n\nSummarises fix: shout hello"))
	})

	It("answers 409 with the conflicting paths and leaves main untouched", func() {
		before := pgCommit(f.repo, "README.md", "hi there\n", "chore: reword readme")

		code, response := merge(map[string]any{"branch": "solo", "mode": "squash", "message": "fix: louder hello"})

		Expect(code).To(Equal(http.StatusConflict), "%v", response)
		Expect(response["conflicts"]).To(Equal([]any{"README.md"}))
		Expect(pgGit(f.repo, "rev-parse", "HEAD")).To(Equal(before))
		Expect(pgGit(f.repo, "status", "--porcelain")).To(BeEmpty())
		Expect(pgGit(f.repo, "branch", "--list", "solo")).NotTo(BeEmpty())
	})

	It("refuses when the main checkout is not on the base branch", func() {
		pgGit(f.repo, "checkout", "-q", "-b", "elsewhere")

		code, response := merge(map[string]any{"branch": "solo", "mode": "squash", "message": "fix: louder hello"})

		Expect(code).To(Equal(http.StatusConflict), "%v", response)
		Expect(response["error"]).To(And(ContainSubstring(`"elsewhere"`), ContainSubstring(`"main"`)))
	})

	DescribeTable("rejects a bad request", func(body map[string]any, status int) {
		code, response := merge(body)
		Expect(code).To(Equal(status), "%v", response)
	},
		Entry("unknown mode", map[string]any{"branch": "solo", "mode": "rebase"}, http.StatusBadRequest),
		Entry("unknown branch", map[string]any{"branch": "nope", "mode": "squash"}, http.StatusNotFound),
		Entry("unknown field", map[string]any{"branch": "solo", "mode": "squash", "cleanup": false}, http.StatusBadRequest),
	)

	It("opens a PR for a branch through prcreate", func() {
		recorder := serveProjectGit(server, http.MethodPost, "/api/projects/acme/branch/pr", map[string]any{"branch": "solo", "draft": true})

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		response := decodeProjectGit[projectBranchPRResponse](recorder)
		Expect(response.Number).To(Equal(7))
		Expect(response.URL).To(Equal("https://github.com/acme/acme/pull/7"))
		Expect(response.TopicBranch).To(HavePrefix("fix/shout-"))
		Expect(created).To(HaveLen(1))
		Expect(created[0]).To(Equal(model.CreatePRInput{
			Title: "fix: shout the greeting", Body: "Summarises fix: shout hello", Head: response.TopicBranch, Base: "main", Draft: true,
		}))
		topicHead := pgGit(f.bare, "rev-parse", "refs/heads/"+response.TopicBranch)
		Expect(pgGit(f.bare, "log", "--format=%s", "main.."+topicHead)).To(Equal("fix: shout hello"))
		Expect(pgGit(f.repo, "rev-parse", "HEAD")).To(Equal(f.base), "opening a PR leaves the checkout alone")
	})

	It("refuses a PR for a branch with nothing past the base", func() {
		pgGit(f.repo, "branch", "empty", f.base)

		recorder := serveProjectGit(server, http.MethodPost, "/api/projects/acme/branch/pr", map[string]any{"branch": "empty"})

		Expect(recorder.Code).To(Equal(http.StatusConflict), recorder.Body.String())
		Expect(created).To(BeEmpty())
	})
})

var _ = Describe("project git OpenAPI", func() {
	It("documents the git summary, branch review and landing routes", func() {
		paths := projectsOpenAPI()["paths"].(map[string]any)
		for path, method := range map[string]string{
			"/api/projects/git-summary":         "get",
			"/api/projects/{name}/git":          "get",
			"/api/projects/{name}/branch/files": "get",
			"/api/projects/{name}/branch/diff":  "get",
			"/api/projects/{name}/branch/merge": "post",
			"/api/projects/{name}/branch/pr":    "post",
		} {
			Expect(paths).To(HaveKey(path))
			Expect(paths[path]).To(HaveKey(method), path)
		}
		statusOp := paths["/api/projects/{name}/status"].(map[string]any)["get"].(map[string]any)
		names := []string{}
		for _, param := range statusOp["parameters"].([]any) {
			names = append(names, param.(map[string]any)["name"].(string))
		}
		Expect(strings.Join(names, ",")).To(ContainSubstring("worktree"))
	})
})
