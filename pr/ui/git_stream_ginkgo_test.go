package ui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	gavelctx "github.com/flanksource/gavel/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("git change push", func() {
	var (
		f       *projectGitFixture
		tracked trackedGitServer
	)

	BeforeEach(func() {
		f = newProjectGitFixture()
		tracked = newTrackedGitServer()
	})

	Describe("GET /api/git/stream", func() {
		It("sends {project: generation} and a higher generation once a touched worktree changed", func() {
			pool, err := tracked.db.DB()
			Expect(err).NotTo(HaveOccurred())
			listener := tracked.server.GitChangeListener(pool)
			listenCtx, stopListening := context.WithCancel(tracked.ctx)
			stopped := make(chan error, 1)
			go func() { stopped <- listener.Run(listenCtx) }()
			DeferCleanup(func() {
				stopListening()
				Eventually(stopped).WithTimeout(10*time.Second).Should(Receive(MatchError(context.Canceled)), "the feed ends only with its context")
			})
			Eventually(listener.Ready()).WithTimeout(10 * time.Second).Should(BeClosed())
			_, err = tracked.tracker.Track(tracked.ctx, f.repo)
			Expect(err).NotTo(HaveOccurred())

			httpServer := httptest.NewServer(tracked.server.Handler())
			DeferCleanup(httpServer.Close)
			resp, err := http.Get(httpServer.URL + "/api/git/stream")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(resp.Body.Close)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			frames := sseFrames(resp)
			var frame string
			Eventually(frames).WithTimeout(10 * time.Second).Should(Receive(&frame))
			first := decodeGenerations(frame)
			Expect(first).To(HaveKey("acme"))
			Expect(first).To(HaveLen(1), "only the configured project is listed")

			Expect(os.WriteFile(filepath.Join(f.worktree, "c.txt"), []byte("new\n"), 0o644)).To(Succeed())
			tracked.tracker.Touch(f.worktree)

			Eventually(frames).WithTimeout(10 * time.Second).Should(Receive(&frame))
			Expect(decodeGenerations(frame)["acme"]).To(BeNumerically(">", first["acme"]))
		})

		It("answers 503 without a git state tracker", func() {
			recorder := serveProjectGit(&Server{ctx: gavelctx.New(context.Background())}, http.MethodGet, "/api/git/stream", nil)
			Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable))
			Expect(decodeProjectGit[map[string]string](recorder)["error"]).To(Equal(gavelctx.ErrNoGitTracker.Error()))
		})
	})

	Describe("POST /api/git/focus", func() {
		DescribeTable("leases the hot cadence to a project's worktree",
			func(body func() map[string]any, want int) {
				recorder := serveProjectGit(tracked.server, http.MethodPost, "/api/git/focus", body())
				Expect(recorder.Code).To(Equal(want), recorder.Body.String())
			},
			Entry("the project directory", func() map[string]any { return map[string]any{"project": "acme"} }, http.StatusNoContent),
			Entry("a linked worktree", func() map[string]any { return map[string]any{"project": "acme", "worktree": f.worktree} }, http.StatusNoContent),
			Entry("an unknown project", func() map[string]any { return map[string]any{"project": "nope"} }, http.StatusNotFound),
			Entry("a path outside the project's worktrees", func() map[string]any {
				return map[string]any{"project": "acme", "worktree": GinkgoT().TempDir()}
			}, http.StatusBadRequest),
			Entry("an unknown field", func() map[string]any { return map[string]any{"project": "acme", "path": f.repo} }, http.StatusBadRequest),
		)

		It("answers 503 without a git state tracker", func() {
			recorder := serveProjectGit(&Server{ctx: gavelctx.New(context.Background())}, http.MethodPost, "/api/git/focus", map[string]any{"project": "acme"})
			Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable))
		})
	})

	Describe("proc status git changes", func() {
		It("counts the primary checkout's uncommitted files from the stored state, without git once tracked", func() {
			Expect(os.WriteFile(filepath.Join(f.repo, "untracked.txt"), []byte("x\n"), 0o644)).To(Succeed())
			projects := []Project{{Name: "acme", Dir: f.repo}, {Name: "plain", Dir: GinkgoT().TempDir()}}
			ctx := tracked.server.context()

			Expect(gitChangeCounts(ctx, projects)).To(Equal(map[string]int{f.repo: 1}), "a directory that is no git work tree has no count")

			withoutGit()
			Expect(procStatusByKey(ctx, projects[:1])["acme"].GitChanges).To(Equal(1))
		})
	})
})
