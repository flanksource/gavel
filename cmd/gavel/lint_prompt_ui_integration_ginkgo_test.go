package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/captain/pkg/aimock"
	"github.com/flanksource/captain/pkg/aimock/anthropicmock"
	"github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons-db/dbtest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prompt lint UI completion", Ordered, func() {
	var dsn string
	BeforeAll(func(ctx SpecContext) {
		handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_lint_prompt_ui"})
		dsn = handle.DSN()
		db, err := database.Open(ctx, database.WithDSN(dsn), database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(db.Close)
	})

	It("saves prompt findings while serving and prints them on shutdown", func() {
		scenario, err := aimock.Parse([]byte(`anthropic:
  - respond:
      text: '{"violations":[{"file":"proposal.md","line":1,"message":"State the decision criteria","severity":"error"}]}'
`))
		Expect(err).NotTo(HaveOccurred())
		server, err := anthropicmock.Start(anthropicmock.Options{Scenario: scenario})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(server.Close)
		root := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(root, "proposal.md"), []byte("# Proposal\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "review.prompt"), []byte("Review the proposal"), 0o600)).To(Succeed())
		bin, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		cmd := exec.Command(bin, "-test.run=^TestLintPromptCLIChild$", "--", "lint", "prompt", "--ui", "--addr", "localhost", "--no-progress", "--", "review.prompt", "proposal.md", "--model", "api:sonnet", "--api-key", aimock.DummyKey, "--api-url", server.APIURL(), "--no-cache", "--bare")
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GAVEL_LINT_PROMPT_CHILD=1", "GAVEL_DB_DSN="+dsn, "CAPTAIN_SESSION_DB_URL="+dsn, "CAPTAIN_DB_CONTEXT=", "CAPTAIN_DB_CONTEXTS=", "GAVEL_ISSUE_ID=")
		checkLintUI(lintUICheck{Command: cmd, File: "proposal.md", ExitCode: 1})
		Expect(server.Remaining()).To(BeEmpty())
	})
})
