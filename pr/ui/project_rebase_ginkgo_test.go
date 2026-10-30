package ui

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/git/branchrebase"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project rebase", func() {
	var f *projectGitFixture
	var server *Server
	BeforeEach(func() { f = newProjectGitFixture(); server = newTrackedGitServer().server })

	It("lists all targets and rebases the specified dirty worktree while the primary checkout is on another branch", func() {
		pgGit(f.repo, "checkout", "-q", "solo")
		response := serveProjectGit(server, http.MethodGet, "/api/projects/acme/branch/rebase?worktree="+url.QueryEscape(f.worktree), nil)
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		ctx := decodeProjectGit[branchrebase.Context](response)
		Expect(ctx.DirtyCount).To(Equal(1))
		Expect(ctx.Targets).To(ContainElement(branchrebase.Target{Ref: "refs/heads/main", Label: "main"}))
		response = serveProjectGit(server, http.MethodPost, "/api/projects/acme/branch/rebase", projectRebaseRequest{Worktree: f.worktree, Branch: pgFeature, Target: "refs/heads/solo", ExpectedHead: f.featureHead})
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(decodeProjectGit[branchrebase.Operation](response).Phase).To(Equal("done"))
		contents, err := os.ReadFile(filepath.Join(f.worktree, "b.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("b\nmore\n"))
		Expect(pgGit(f.repo, "branch", "--show-current")).To(Equal("solo"))
	})

	It("rejects paths outside the project's worktrees", func() {
		response := serveProjectGit(server, http.MethodPost, "/api/projects/acme/branch/rebase", projectRebaseRequest{Worktree: f.bare, Branch: pgFeature, Target: "refs/heads/main", ExpectedHead: f.featureHead})
		Expect(response.Code).To(Equal(http.StatusBadRequest))
	})

	It("rejects unknown fields without mutating the repository", func() {
		response := serveProjectGit(server, http.MethodPost, "/api/projects/acme/branch/rebase", map[string]any{"worktree": f.worktree, "resolved": true})
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(pgGit(f.worktree, "rev-parse", "HEAD")).To(Equal(f.featureHead))
	})

	It("returns recoverable conflict state and supports abort through the API", func() {
		newHead := pgCommit(f.worktree, "README.md", "topic\n", "topic conflict")
		pgCommit(f.repo, "README.md", "main\n", "main conflict")
		response := serveProjectGit(server, http.MethodPost, "/api/projects/acme/branch/rebase", projectRebaseRequest{Worktree: f.worktree, Branch: pgFeature, Target: "refs/heads/main", ExpectedHead: newHead})
		Expect(response.Code).To(Equal(http.StatusConflict), response.Body.String())
		failure := decodeProjectGit[struct {
			Operation branchrebase.Operation `json:"operation"`
		}](response)
		Expect(failure.Operation.Conflicts).To(Equal([]string{"README.md"}))
		response = serveProjectGit(server, http.MethodPost, "/api/projects/acme/branch/rebase/abort", projectRebaseRequest{Worktree: f.worktree, Operation: failure.Operation.ID})
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(pgGit(f.worktree, "rev-parse", "HEAD")).To(Equal(newHead))
	})
})
