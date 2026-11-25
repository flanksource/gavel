package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

func TestLintUIChild(t *testing.T) {
	if os.Getenv("GAVEL_LINT_UI_CHILD") != "1" {
		return
	}
	os.Args = append([]string{"gavel"}, os.Args[3:]...)
	main()
}

var _ = Describe("Lint UI completion", func() {
	DescribeTable("persists completed runs before shutdown and prints the latest verdict", func(initial, rerun string, wantExit int) {
		root := GinkgoT().TempDir()
		file := filepath.Join(root, "settings.json")
		Expect(os.WriteFile(file, []byte(initial), 0o600)).To(Succeed())
		bin, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		cmd := exec.Command(bin, "-test.run=^TestLintUIChild$", "--", "lint", "json", "settings.json", "--ui", "--addr", "localhost", "--no-progress", "--no-lint-cache")
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GAVEL_LINT_UI_CHILD=1")
		checkLintUI(lintUICheck{Command: cmd, File: "settings.json", Rerun: rerun, ExitCode: wantExit})
	},
		Entry("initial findings", `{"enabled":`, "", 1),
		Entry("clean rerun clears the initial failure", `{"enabled":`, `{"enabled":true}`, 0),
		Entry("failed rerun replaces the initial clean verdict", `{"enabled":true}`, `{"enabled":`, 1),
	)
})

type lintUICheck struct {
	Command  *exec.Cmd
	File     string
	Rerun    string
	ExitCode int
}

func checkLintUI(check lintUICheck) {
	GinkgoHelper()
	cmd := check.Command
	output := gbytes.NewBuffer()
	cmd.Stdout, cmd.Stderr = output, output
	Expect(cmd.Start()).To(Succeed())
	finished := make(chan error, 1)
	go func() { finished <- cmd.Wait() }()
	DeferCleanup(func() { _ = cmd.Process.Kill() })
	urlPattern := regexp.MustCompile(`Test UI at (http://localhost:[0-9]+)`)
	var url string
	Eventually(func() string {
		if match := urlPattern.FindSubmatch(output.Contents()); len(match) > 1 {
			url = string(match[1])
		}
		return url
	}, 10*time.Second).ShouldNot(BeEmpty(), string(output.Contents()))
	client := &http.Client{Timeout: time.Second}
	Eventually(func(g Gomega) {
		response, err := client.Get(url + "/api/tests")
		g.Expect(err).NotTo(HaveOccurred())
		defer response.Body.Close()
		var current testui.Snapshot
		g.Expect(json.NewDecoder(response.Body).Decode(&current)).To(Succeed())
		g.Expect(current.Status.Running).To(BeFalse())
		g.Expect(current.Lint).To(HaveLen(1))
	}, 10*time.Second).Should(Succeed())
	runs, err := filepath.Glob(filepath.Join(cmd.Dir, ".gavel", "run-*.json"))
	Expect(err).NotTo(HaveOccurred())
	Expect(runs).To(HaveLen(1), "the initial lint run must be saved while the UI is still serving")
	if check.Rerun != "" {
		Expect(os.WriteFile(filepath.Join(cmd.Dir, check.File), []byte(check.Rerun), 0o600)).To(Succeed())
		response, err := client.Post(url+"/api/rerun", "application/json", strings.NewReader(`{"lint":true,"lint_files":["settings.json"],"lint_linters":["json"]}`))
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Body.Close()).To(Succeed())
		Expect(response.StatusCode).To(Equal(http.StatusAccepted))
		runs, err = filepath.Glob(filepath.Join(cmd.Dir, ".gavel", "run-*.json"))
		Expect(err).NotTo(HaveOccurred())
		Expect(runs).To(HaveLen(2), "each completed UI rerun must have its own history entry")
	}
	latest, err := snapshots.LatestRuns(cmd.Dir)
	Expect(err).NotTo(HaveOccurred())
	Expect(latest).To(HaveLen(1))
	saved, err := snapshots.LoadRunFile(latest[0])
	Expect(err).NotTo(HaveOccurred())
	Expect(saved.Metadata.Kind).To(Equal("lint"))
	Expect(saved.Status.Running).To(BeFalse())
	Expect(saved.Lint).To(HaveLen(1))
	Expect(cmd.Process.Signal(os.Interrupt)).To(Succeed())
	Eventually(finished, 10*time.Second).Should(Receive(&err))
	Expect(cmd.ProcessState.ExitCode()).To(Equal(check.ExitCode), string(output.Contents()))
	if check.ExitCode == 0 {
		Expect(err).NotTo(HaveOccurred())
		Expect(saved.Lint[0].Violations).To(BeEmpty())
	} else {
		Expect(err).To(HaveOccurred())
		Expect(saved.Lint[0].Violations).NotTo(BeEmpty())
		Expect(saved.Lint[0].Violations[0].Message).NotTo(BeNil())
		Expect(string(output.Contents())).To(ContainSubstring(*saved.Lint[0].Violations[0].Message))
		Expect(string(output.Contents())).To(ContainSubstring(check.File))
	}
}
