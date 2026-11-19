package ui

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project checkout switch", func() {
	var f *projectGitFixture
	var server *Server
	var landed string

	BeforeEach(func() {
		f = newProjectGitFixture()
		server = &Server{}
		landed = pgCommit(f.repo, "CHANGELOG.md", "v2\n", "docs: changelog")
		pgGit(f.repo, "push", "-q", "origin", "main")
		pgGit(f.repo, "reset", "-q", "--hard", f.base)
		pgGit(f.repo, "checkout", "-q", "solo")
	})

	switchTo := func(sync string) (int, map[string]any) {
		GinkgoHelper()
		recorder := serveProjectGit(server, http.MethodPost, "/api/projects/acme/checkout/switch", map[string]any{
			"branch": "main", "expectedHead": f.soloHead, "sync": sync,
		})
		return recorder.Code, decodeProjectGit[map[string]any](recorder)
	}

	It("plans the switch from the last-fetched upstream", func() {
		recorder := serveProjectGit(server, http.MethodGet, "/api/projects/acme/checkout/switch?branch=main", nil)
		plan := decodeProjectGit[map[string]any](recorder)

		Expect(recorder.Code).To(Equal(http.StatusOK), "%v", plan)
		Expect(plan).To(Equal(map[string]any{
			"repo": f.repo, "current": "solo", "head": f.soloHead, "branch": "main", "branchHead": f.base,
			"upstream": "origin/main", "upstreamHead": landed, "ahead": float64(0), "behind": float64(1),
			"unpreserved": []any{}, "dirtyCount": float64(0),
		}))
	})

	It("fast-forwards main to origin/main and checks it out", func() {
		code, response := switchTo("fast-forward")

		Expect(code).To(Equal(http.StatusOK), "%v", response)
		Expect(response).To(Equal(map[string]any{
			"branch": "main", "head": landed, "sync": "fast-forward",
			"previousBranch": "solo", "previousHead": f.soloHead, "previousBranchHead": f.base,
		}))
		Expect(pgGit(f.repo, "branch", "--show-current")).To(Equal("main"))
	})

	It("refuses with 409 a reset that would orphan a commit only main holds", func() {
		pgGit(f.repo, "checkout", "-q", "main")
		local := pgCommit(f.repo, "local.txt", "mine\n", "feat: local only")
		pgGit(f.repo, "checkout", "-q", "solo")

		code, response := switchTo("reset")

		Expect(code).To(Equal(http.StatusConflict))
		Expect(response["error"]).To(ContainSubstring(local[:8] + " feat: local only"))
		Expect(pgGit(f.repo, "rev-parse", "main")).To(Equal(local))
	})

	It("rejects an unknown sync mode with 400", func() {
		code, _ := switchTo("merge")

		Expect(code).To(Equal(http.StatusBadRequest))
	})
})
