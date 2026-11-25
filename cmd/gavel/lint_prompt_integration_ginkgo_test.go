package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/aimock"
	"github.com/flanksource/captain/pkg/aimock/anthropicmock"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons-db/dbtest"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/workspace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLintPromptCLIChild(t *testing.T) {
	if os.Getenv("GAVEL_LINT_PROMPT_CHILD") != "1" {
		return
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	captainconfig.SetPathForTesting(filepath.Join(dir, "captain.yaml"))
	os.Args = append([]string{"gavel"}, os.Args[3:]...)
	main()
}

var _ = Describe("Lint prompt CLI integration", Ordered, func() {
	var dsn string
	BeforeAll(func(ctx SpecContext) {
		handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_lint_prompt"})
		dsn = handle.DSN()
		db, err := database.Open(ctx, database.WithDSN(dsn), database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(db.Close)
	})

	DescribeTable("returns lint snapshots with the standard exit verdict", func(output string, exitCode, findings int, success bool) {
		reply, err := json.Marshal(output)
		Expect(err).NotTo(HaveOccurred())
		scenario, err := aimock.Parse([]byte(fmt.Sprintf("anthropic:\n  - respond: {text: %s}\n", reply)))
		Expect(err).NotTo(HaveOccurred())
		server, err := anthropicmock.Start(anthropicmock.Options{Scenario: scenario})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(server.Close)
		bin, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		cmd := exec.Command(bin, "-test.run=^TestLintPromptCLIChild$", "--", "lint", "prompt", "--json", "--no-progress", "--exclude", "prompt:excluded", "--", "proposal-review.prompt", "--model", "api:sonnet", "--api-key", aimock.DummyKey, "--api-url", server.APIURL(), "--no-cache", "--bare")
		cmd.Dir = GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(cmd.Dir, "proposal-review.prompt"), []byte("Review the change"), 0o600)).To(Succeed())
		cmd.Env = append(os.Environ(), "GAVEL_LINT_PROMPT_CHILD=1", "GAVEL_DB_DSN="+dsn, "CAPTAIN_SESSION_DB_URL="+dsn, "CAPTAIN_DB_CONTEXT=", "CAPTAIN_DB_CONTEXTS=", "GAVEL_ISSUE_ID=")
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		err = cmd.Run()
		if exitCode == 0 {
			Expect(err).NotTo(HaveOccurred(), stdout.String()+stderr.String())
		} else {
			var failed *exec.ExitError
			Expect(err).To(BeAssignableToTypeOf(failed), stderr.String())
			Expect(cmd.ProcessState.ExitCode()).To(Equal(exitCode), stderr.String())
		}
		var snapshot testui.Snapshot
		Expect(stderr.String()).NotTo(ContainSubstring("lint failed"))
		Expect(json.Unmarshal(stdout.Bytes(), &snapshot)).To(Succeed(), stdout.String()+stderr.String())
		Expect(snapshot.Lint).To(HaveLen(1))
		Expect(snapshot.Lint[0].Linter).To(Equal("proposal-review"))
		Expect(snapshot.Lint[0].Success).To(Equal(success), stderr.String())
		Expect(snapshot.Lint[0].Violations).To(HaveLen(findings))
		Expect(snapshot.Metadata.Kind).To(Equal("lint"))
		Expect(server.Remaining()).To(BeEmpty())
		requests, err := json.Marshal(server.Requests())
		Expect(err).NotTo(HaveOccurred())
		Expect(string(requests)).To(ContainSubstring(`violations`))
	},
		Entry("clean", `{"violations":[]}`, 0, 0, true),
		Entry("error finding", `{"violations":[{"file":"service.go","line":12,"message":"Return the failure","severity":"error"}]}`, 1, 1, true),
		Entry("excluded error finding", `{"violations":[{"file":"service.go","message":"Excluded rule","rule":{"method":"excluded"},"severity":"error"}]}`, 0, 0, true),
		Entry("warning finding", `{"violations":[{"message":"Review the edge case","severity":"warning"}]}`, 0, 1, true),
		Entry("invalid schema", `{"unrelated":true}`, 1, 0, false),
	)

	// childRun runs `gavel lint prompt` in a child process over Files written
	// under Root (a fresh temp dir when empty). Dir is the directory the process
	// runs in, relative to Root unless absolute; Global are gavel flags placed
	// before the subcommand.
	type childRun struct {
		Replies []string
		Files   map[string]string
		Args    []string
		Paths   []string
		Global  []string
		Dir     string
		Root    string
		Prompt  string
	}
	runChild := func(run childRun) (*anthropicmock.Server, testui.Snapshot, string) {
		var scenario bytes.Buffer
		scenario.WriteString("anthropic:\n")
		for _, reply := range run.Replies {
			quoted, err := json.Marshal(reply)
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(&scenario, "  - respond: {text: %s}\n", quoted)
		}
		parsed, err := aimock.Parse(scenario.Bytes())
		Expect(err).NotTo(HaveOccurred())
		server, err := anthropicmock.Start(anthropicmock.Options{Scenario: parsed})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(server.Close)
		bin, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		if run.Root == "" {
			run.Root = GinkgoT().TempDir()
		}
		if run.Prompt == "" {
			run.Prompt = "review.prompt"
		}
		base := append(append([]string{"-test.run=^TestLintPromptCLIChild$", "--"}, run.Global...), "lint", "prompt", "--json", "--no-progress")
		base = append(append(base, run.Args...), "--", run.Prompt, "--model", "api:sonnet", "--api-key", aimock.DummyKey, "--api-url", server.APIURL(), "--no-cache", "--bare")
		cmd := exec.Command(bin, append(base, run.Paths...)...)
		cmd.Dir = run.Dir
		if !filepath.IsAbs(cmd.Dir) {
			cmd.Dir = filepath.Join(run.Root, run.Dir)
		}
		run.Files["review.prompt"] = `Review {{join files " "}}`
		for path, content := range run.Files {
			Expect(os.MkdirAll(filepath.Dir(filepath.Join(run.Root, path)), 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(run.Root, path), []byte(content), 0o600)).To(Succeed())
		}
		cmd.Env = append(append(os.Environ(), server.Env()...), "HOME="+run.Root, "GAVEL_LINT_PROMPT_CHILD=1", "GAVEL_DB_DSN="+dsn, "CAPTAIN_SESSION_DB_URL="+dsn, "CAPTAIN_DB_CONTEXT=", "CAPTAIN_DB_CONTEXTS=", "GAVEL_ISSUE_ID=")
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		Expect(cmd.Run()).To(Succeed(), stdout.String()+stderr.String())
		var snapshot testui.Snapshot
		Expect(json.Unmarshal(stdout.Bytes(), &snapshot)).To(Succeed(), stdout.String()+stderr.String())
		requests, err := json.Marshal(server.Requests())
		Expect(err).NotTo(HaveOccurred())
		return server, snapshot, string(requests)
	}

	DescribeTable("reviews every path argument in one run with the paths as the files list", func(paths []string, rendered string, enum bool) {
		server, snapshot, requests := runChild(childRun{
			Replies: []string{`{"violations":[]}`},
			Files:   map[string]string{"pkg/service.go": "package pkg\n", "cmd/main.go": "package main\n"},
			Paths:   paths,
		})
		Expect(server.Remaining()).To(BeEmpty())
		Expect(snapshot.Lint).To(HaveExactElements(SatisfyAll(HaveField("Linter", "review"), HaveField("Success", true), HaveField("RunID", Not(BeEmpty())))))
		Expect(server.Requests()).To(HaveLen(1))
		Expect(requests).To(ContainSubstring(rendered))
		Expect(requests).To(SatisfyAll(ContainSubstring(`"read_file"`), ContainSubstring("report_violation")), "the reviewing agent is offered gavel's tools")
		enumOf, err := json.Marshal(map[string]any{"enum": paths})
		Expect(err).NotTo(HaveOccurred())
		if enum {
			Expect(requests).To(ContainSubstring(string(enumOf[1:len(enumOf)-1])), "report_violation's file is one of the files under review")
		} else {
			Expect(requests).NotTo(ContainSubstring(`"enum":["pkg"`))
		}
	},
		Entry("files", []string{"pkg/service.go", "cmd/main.go"}, "Review pkg/service.go cmd/main.go", true),
		Entry("a directory", []string{"pkg", "cmd/main.go"}, "Review pkg cmd/main.go", false),
	)

	It("repairs findings with lint.fix and re-runs the prompt until clean", func() {
		server, snapshot, requests := runChild(childRun{
			Replies: []string{`{"violations":[{"file":"pkg/service.go","line":1,"message":"Return the failure","severity":"error"}]}`, "Fixed pkg/service.go", `{"violations":[]}`},
			Files:   map[string]string{"pkg/service.go": "package pkg\n", ".gavel.yaml": "lint:\n  fix:\n    model: api:sonnet\n"},
			Args:    []string{"--ai-fix", "--ai-fix-max-iterations", "1"},
			Paths:   []string{"pkg"},
		})
		Expect(server.Remaining()).To(BeEmpty())
		Expect(snapshot.Lint).To(HaveExactElements(SatisfyAll(HaveField("Success", true), HaveField("Violations", BeEmpty()))))
		Expect(requests).To(ContainSubstring("Return the failure"))
	})

	// gitRepo is a fresh git working tree with a pkg/ subdirectory and no .gavel.
	gitRepo := func() string {
		root, err := filepath.EvalSymlinks(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		Expect(exec.Command("git", "init", "-q", root).Run()).To(Succeed())
		return root
	}
	expectRunUnderRoot := func(root string) {
		runs, err := filepath.Glob(filepath.Join(root, ".gavel", "run-*.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(runs).To(HaveLen(1), "the run is recorded in the git root's .gavel")
		Expect(filepath.Join(root, "pkg", ".gavel")).NotTo(BeADirectory(), "no .gavel is created in the subdirectory")
		dirs, err := workspace.Resolve(root)
		Expect(err).NotTo(HaveOccurred())
		latest, err := latestRunFiles(dirs.StateRoot)
		Expect(err).NotTo(HaveOccurred())
		Expect(latest).To(Equal(runs), "gavel ui serve from the root finds the run")
	}

	It("records a run started in a repo subdirectory under the git root's .gavel", func() {
		root := gitRepo()
		runChild(childRun{
			Root: root, Dir: "pkg",
			Replies: []string{`{"violations":[]}`},
			Files:   map[string]string{"pkg/service.go": "package pkg\n", "pkg/review.prompt": `Review {{join files " "}}`},
			Paths:   []string{"service.go"},
		})
		expectRunUnderRoot(root)
	})

	It("resolves --cwd for the run's directories, not the process working directory", func() {
		root := gitRepo()
		outside := GinkgoT().TempDir()
		runChild(childRun{
			Root: root, Dir: outside, Global: []string{"--cwd", filepath.Join(root, "pkg")},
			Replies: []string{`{"violations":[]}`},
			Files:   map[string]string{"pkg/service.go": "package pkg\n"},
			Prompt:  filepath.Join(root, "review.prompt"),
			Paths:   []string{"service.go"},
		})
		Expect(filepath.Join(outside, ".gavel")).NotTo(BeADirectory())
		expectRunUnderRoot(root)
	})
})
