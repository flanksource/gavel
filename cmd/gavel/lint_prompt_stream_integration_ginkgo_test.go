package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/flanksource/captain/pkg/aimock"
	"github.com/flanksource/captain/pkg/aimock/anthropicmock"
	"github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Prompt lint live reporting", Ordered, func() {
	var dsn string
	BeforeAll(func(ctx SpecContext) {
		handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_lint_prompt_stream"})
		dsn = handle.DSN()
		db, err := database.Open(ctx, database.WithDSN(dsn), database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(db.Close)
	})

	DescribeTable("writes accepted findings to stderr and one snapshot before the model finishes", func(ui, interrupt bool) {
		scenario, err := aimock.Parse([]byte(`anthropic:
  - respond:
      tool_use:
        name: report_violation
        input:
          violations:
            - {file: proposal.md, line: 1, severity: error, message: State the decision criteria}
            - {file: proposal.md, line: 1, severity: error, message: State the decision criteria}
            - {file: proposal.md, line: 1, severity: error, message: Excluded finding, rule: {method: excluded}}
      stop_reason: tool_use
  - respond:
      tool_use:
        name: lint
        input: {paths: [settings.json], linters: [json]}
      stop_reason: tool_use
  - respond:
      text: '{"violations":[{"file":"proposal.md","line":1,"severity":"error","message":"State the decision criteria"},{"file":"proposal.md","line":1,"severity":"warning","message":"Explain the alternative"}]}'
`))
		Expect(err).NotTo(HaveOccurred())
		mock, err := anthropicmock.Start(anthropicmock.Options{Scenario: scenario})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(mock.Close)
		target, err := url.Parse(mock.APIURL())
		Expect(err).NotTo(HaveOccurred())
		proxy := httputil.NewSingleHostReverseProxy(target)
		blocked, release := make(chan int32, 2), make(chan struct{})
		var calls atomic.Int32
		gate := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			if request.URL.Path == "/v1/messages" && calls.Add(1) > 1 {
				blocked <- calls.Load()
				select {
				case <-release:
				case <-request.Context().Done():
					return
				}
			}
			proxy.ServeHTTP(w, request)
		}))
		DeferCleanup(gate.Close)
		DeferCleanup(func() { close(release) })
		root := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(root, "proposal.md"), []byte("# Proposal\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "settings.json"), []byte(`{"enabled":`), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "review.prompt"), []byte("Review the proposal"), 0o600)).To(Succeed())
		bin, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		args := []string{"-test.run=^TestLintPromptCLIChild$", "--", "lint", "prompt", "--json", "--no-progress", "--exclude", "prompt:excluded"}
		if ui {
			args = append(args, "--ui", "--addr", "localhost")
		}
		args = append(args, "--", "review.prompt", "proposal.md", "--model", "api:sonnet", "--api-key", aimock.DummyKey, "--api-url", gate.URL, "--no-cache", "--bare")
		cmd := exec.Command(bin, args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GAVEL_LINT_PROMPT_CHILD=1", "GAVEL_DB_DSN="+dsn, "CAPTAIN_SESSION_DB_URL="+dsn, "CAPTAIN_DB_CONTEXT=", "CAPTAIN_DB_CONTEXTS=", "GAVEL_ISSUE_ID=")
		stdout, stderr := gbytes.NewBuffer(), gbytes.NewBuffer()
		cmd.Stdout, cmd.Stderr = stdout, stderr
		Expect(cmd.Start()).To(Succeed())
		finished := make(chan error, 1)
		go func() { finished <- cmd.Wait() }()
		DeferCleanup(func() { _ = cmd.Process.Kill() })
		Eventually(blocked, 15*time.Second).Should(Receive(Equal(int32(2))), string(stderr.Contents()))
		Expect(string(stderr.Contents())).To(ContainSubstring("State the decision criteria"))
		Expect(string(stderr.Contents())).NotTo(ContainSubstring("Excluded finding"))
		paths, err := filepath.Glob(filepath.Join(root, ".gavel", "run-*.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(HaveLen(1))
		live, err := snapshots.LoadRunFile(paths[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(live.Status.Running).To(BeTrue())
		Expect(live.Lint).To(HaveLen(1))
		Expect(live.Lint[0].Violations).To(HaveLen(1))
		Expect(finished).NotTo(Receive())
		if interrupt {
			Expect(cmd.Process.Signal(os.Interrupt)).To(Succeed())
		} else {
			release <- struct{}{}
			Eventually(blocked, 15*time.Second).Should(Receive(Equal(int32(3))), string(stderr.Contents()))
			nested, err := snapshots.LoadRunFile(paths[0])
			Expect(err).NotTo(HaveOccurred())
			Expect(nested.Status.Running).To(BeTrue())
			Expect(nested.Lint).To(ContainElement(SatisfyAll(HaveField("Linter", HaveSuffix("/json")), HaveField("Success", true), HaveField("Violations", HaveLen(1)))))
			Expect(string(stderr.Contents())).To(ContainSubstring("unexpected EOF"))
			release <- struct{}{}
			if ui {
				Eventually(func() bool {
					snapshot, err := snapshots.LoadRunFile(paths[0])
					return err == nil && !snapshot.Status.Running
				}, 15*time.Second).Should(BeTrue())
				Expect(cmd.Process.Signal(os.Interrupt)).To(Succeed())
			}
		}
		Eventually(finished, 15*time.Second).Should(Receive(&err))
		Expect(cmd.ProcessState.ExitCode()).To(Equal(1), string(stderr.Contents()))
		Expect(strings.Count(string(stderr.Contents()), "State the decision criteria")).To(Equal(1))
		var final testui.Snapshot
		Expect(json.Unmarshal(stdout.Contents(), &final)).To(Succeed(), string(stdout.Contents())+string(stderr.Contents()))
		Expect(final.Status.Running).To(BeFalse())
		wantFindings := 2
		if interrupt {
			wantFindings = 1
		}
		Expect(final.Lint[0].Violations).To(HaveLen(wantFindings))
		if !interrupt {
			Expect(string(stderr.Contents())).To(ContainSubstring("Explain the alternative"))
		}
		pathsAfter, err := filepath.Glob(filepath.Join(root, ".gavel", "run-*.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(pathsAfter).To(Equal(paths), "a live run is finalized in its existing history file")
	},
		Entry("without UI", false, false),
		Entry("with UI", true, false),
		Entry("interrupted without UI", false, true),
		Entry("interrupted with UI", true, true),
	)
})
