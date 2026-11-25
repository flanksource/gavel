package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lint run reports", func() {
	message := "statement failed"
	results := []*linters.LinterResult{{Linter: "oipa-transaction", Success: true,
		Violations: []models.Violation{{File: "tr/Fee/XMLData.xml", Line: 12, Severity: models.SeverityError, Message: &message}}}}
	git := func(dir string, args ...string) string {
		cmd := exec.Command("git", append([]string{"-c", "user.email=lint@example.com", "-c", "user.name=lint"}, args...)...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(out))
		return strings.TrimSpace(string(out))
	}
	read := func(path string) testui.Snapshot {
		raw, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		var snap testui.Snapshot
		Expect(json.Unmarshal(raw, &snap)).To(Succeed())
		return snap
	}

	It("saves a run in a worktree under its HEAD and as its own run", func() {
		dir := GinkgoT().TempDir()
		git(dir, "init", "-q")
		Expect(os.WriteFile(filepath.Join(dir, "README.md"), []byte("rules\n"), 0o600)).To(Succeed())
		git(dir, "add", "README.md")
		git(dir, "commit", "-qm", "chore: init")
		head := git(dir, "rev-parse", "HEAD")

		path, err := SaveSnapshot(dir, Snapshot(SnapshotOptions{WorkDir: dir, Version: "test", Started: time.Now().UTC(), Results: results}))

		Expect(err).NotTo(HaveOccurred())
		Expect(filepath.Base(path)).To(HavePrefix("run-"))
		last, err := snapshots.ResolveLast(dir)
		Expect(err).NotTo(HaveOccurred())
		for _, saved := range []testui.Snapshot{read(path), read(last)} {
			Expect([]any{saved.Metadata.Kind, saved.Status.LintRun, saved.Git.SHA, saved.Lint[0].Violations[0].Line}).
				To(Equal([]any{"lint", true, head, 12}))
		}
	})

	It("saves a run outside a worktree only as its own run", func() {
		dir := GinkgoT().TempDir()

		path, err := SaveSnapshot(dir, Snapshot(SnapshotOptions{WorkDir: dir, Started: time.Now().UTC(), Results: results}))

		Expect(err).NotTo(HaveOccurred())
		Expect(read(path).Lint).To(HaveLen(1))
		Expect(filepath.Join(dir, snapshots.Dir, "last.json")).NotTo(BeAnExistingFile())
	})

	It("serves the results until its context ends", func() {
		var banner bytes.Buffer
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan testui.Snapshot, 1)
		go func() {
			defer GinkgoRecover()
			snap, err := Serve(ctx, ServeOptions{Addr: "127.0.0.1", Writer: &banner, Results: results})
			Expect(err).NotTo(HaveOccurred())
			done <- snap
		}()
		Eventually(banner.String).Should(HavePrefix("Lint UI at http://"))
		url := strings.TrimSpace(strings.TrimPrefix(banner.String(), "Lint UI at "))

		var served testui.Snapshot
		Eventually(func() error {
			resp, err := http.Get(url + "/api/tests")
			if err != nil {
				return err
			}
			defer resp.Body.Close() //nolint:errcheck
			return json.NewDecoder(resp.Body).Decode(&served)
		}).Should(Succeed())
		Expect(served.Lint).To(HaveExactElements(HaveField("Linter", "oipa-transaction")))

		cancel()
		Eventually(done).Should(Receive(HaveField("Lint", HaveLen(1))))
	})
})
