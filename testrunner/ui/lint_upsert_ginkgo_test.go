package testui_test

import (
	"github.com/flanksource/gavel/linters"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server.UpsertLintResult", func() {
	const workDir = "/repo"

	It("shows pending entries and keeps the run going until every entry is final", func() {
		srv := testui.NewServer()
		srv.BeginRun("initial")
		srv.UpsertLintResult(&linters.LinterResult{Linter: "review:a.go", WorkDir: workDir, Status: linters.LinterStatusPending})
		srv.UpsertLintResult(&linters.LinterResult{Linter: "review:b.go", WorkDir: workDir, Status: linters.LinterStatusPending})
		srv.MarkDone()

		snap := srv.Snapshot()
		Expect(snap.Status.LintRun).To(BeTrue())
		Expect(snap.Status.Running).To(BeTrue(), "pending lint entries keep the run open")
		Expect(snap.Lint).To(HaveLen(2))

		srv.UpsertLintResult(&linters.LinterResult{Linter: "review:a.go", WorkDir: workDir, Status: linters.LinterStatusRunning, RunID: "run-a"})
		srv.UpsertLintResult(&linters.LinterResult{Linter: "review:b.go", WorkDir: workDir, Success: true})
		snap = srv.Snapshot()
		Expect(snap.Status.Running).To(BeTrue())
		Expect(snap.Lint).To(Equal([]*linters.LinterResult{
			{Linter: "review:a.go", WorkDir: workDir, Status: linters.LinterStatusRunning, RunID: "run-a"},
			{Linter: "review:b.go", WorkDir: workDir, Success: true},
		}))

		srv.UpsertLintResult(&linters.LinterResult{Linter: "review:a.go", WorkDir: workDir, Success: true, RunID: "run-a"})
		Expect(srv.Snapshot().Status.Running).To(BeFalse())
	})

	It("serves Captain's prompt runs so a result's run_id can be followed", func() {
		response := get(testui.NewServer().Handler(), "/api/captain/prompt/runs/unknown-run")
		Expect(response.Code).To(Equal(404))
		Expect(response.Body.String()).NotTo(ContainSubstring("<html"), "the route must be Captain's, not the SPA fallback")
	})

	It("keys entries by linter and work dir", func() {
		srv := testui.NewServer()
		srv.UpsertLintResult(&linters.LinterResult{Linter: "review:a.go", WorkDir: "/one", Success: true})
		srv.UpsertLintResult(&linters.LinterResult{Linter: "review:a.go", WorkDir: "/two", Success: true})
		Expect(srv.Snapshot().Lint).To(HaveLen(2))
	})
})
