package ui

import (
	"net/http"
	"net/url"
	"os/exec"

	"github.com/flanksource/gavel/git/branchdelete"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project branch delete", func() {
	var f *projectGitFixture
	var server *Server
	BeforeEach(func() { f = newProjectGitFixture(); server = newTrackedGitServer().server })

	It("plans the worktree's uncommitted files and unmerged commits, refuses without acknowledgement, then deletes both", func() {
		response := serveProjectGit(server, http.MethodGet, "/api/projects/acme/branch/delete?branch="+url.QueryEscape(pgFeature), nil)
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		plan := decodeProjectGit[branchdelete.Plan](response)
		Expect([]any{plan.Head, plan.Worktree, plan.Uncommitted, len(plan.Unmerged)}).To(Equal([]any{f.featureHead, f.worktree, []string{"b.txt"}, 2}))

		request := projectBranchDeleteRequest{Branch: pgFeature, ExpectedHead: plan.Head}
		response = serveProjectGit(server, http.MethodPost, "/api/projects/acme/branch/delete", request)
		Expect(response.Code).To(Equal(http.StatusConflict), response.Body.String())
		Expect(f.worktree).To(BeADirectory())

		request.DiscardUncommitted, request.DeleteUnmerged = true, true
		response = serveProjectGit(server, http.MethodPost, "/api/projects/acme/branch/delete", request)
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		Expect(decodeProjectGit[branchdelete.Result](response)).To(Equal(branchdelete.Result{WorktreeRemoved: true, BranchDeleted: true}))
		Expect(f.worktree).NotTo(BeADirectory())
		Expect(exec.Command("git", "-C", f.repo, "rev-parse", "--verify", "refs/heads/"+pgFeature).Run()).To(HaveOccurred())
	})

	It("refuses to delete the branch the main checkout has out", func() {
		response := serveProjectGit(server, http.MethodGet, "/api/projects/acme/branch/delete?branch=solo", nil)
		Expect(response.Code).To(Equal(http.StatusBadRequest), response.Body.String())
	})
})
