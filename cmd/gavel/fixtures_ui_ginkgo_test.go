package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixtures UI command", func() {
	It("registers the live and detached UI controls in fixture help", func() {
		for _, name := range []string{"ui", "addr", "detach", "auto-stop", "idle-timeout"} {
			Expect(fixturesCmd.Flags().Lookup(name)).NotTo(BeNil(), "missing fixtures --%s", name)
		}
		Expect(fixturesHelp(fixturesCmd).ANSI()).To(ContainSubstring("gavel fixtures --ui checks.fixture.md"))
	})

	It("routes the direct fixture runner through fixture-only test UI orchestration", func() {
		display := &fixtures.DisplayOptions{ShowPassed: true}
		runner := &fixtures.RunnerOptions{
			Paths:        []string{"verification.md"},
			WorkDir:      "/workspace",
			UpdateGolden: true,
			Display:      display,
		}
		ui := fixtureUIOptions{
			UI:          true,
			Addr:        "127.0.0.1",
			Detach:      true,
			AutoStop:    15 * time.Minute,
			IdleTimeout: 2 * time.Minute,
		}

		opts, detach := fixtureUIRunOptions(fixtureUIRunRequest{Runner: runner, UI: ui})

		Expect(detach).To(BeTrue())
		Expect(opts.UI).To(BeTrue())
		Expect(opts.Addr).To(Equal("127.0.0.1"))
		Expect(opts.Fixtures).To(BeTrue())
		Expect(opts.FixturesOnly).To(BeTrue())
		Expect(opts.FixtureRunner).To(BeIdenticalTo(runner))
		Expect(opts.WorkDir).To(Equal("/workspace"))
		Expect(opts.SkipHooks).To(BeTrue())
		Expect(opts.AutoStop).To(Equal(15 * time.Minute))
		Expect(opts.IdleTimeout).To(Equal(2 * time.Minute))
	})

	It("keeps a failed fixture run's UI answering /api/tests and /api/rerun until it is stopped", func() {
		workDir := GinkgoT().TempDir()
		path := filepath.Join(workDir, "failing.fixture.md")
		Expect(os.WriteFile(path, []byte("# Failing\n\n```bash\nexit 3\n```\n"), 0o600)).To(Succeed())
		_, runErr := testrunner.Run(testrunner.RunOptions{WorkDir: workDir, FixturesOnly: true, FixtureRunner: &fixtures.RunnerOptions{Paths: []string{path}, WorkDir: workDir}})
		Expect(runErr).To(HaveOccurred())

		server := testui.NewServer()
		reran := false
		server.SetRerunFunc(func(testui.RerunRequest, *testui.RerunOutputBuffer) error { reran = true; return nil })
		served := httptest.NewServer(server.Handler())
		DeferCleanup(served.Close)
		held := false

		holdFailedRunUI(server, runErr, func() {
			held = true
			var snapshot testui.Snapshot
			response, err := served.Client().Get(served.URL + "/api/tests")
			Expect(err).ToNot(HaveOccurred())
			defer response.Body.Close()
			Expect(json.NewDecoder(response.Body).Decode(&snapshot)).To(Succeed())
			Expect(snapshot.Error).To(Equal(runErr.Error()))
			Expect(snapshot.Status.Running).To(BeFalse())
			rerun, err := served.Client().Post(served.URL+"/api/rerun", "application/json", strings.NewReader(`{}`))
			Expect(err).ToNot(HaveOccurred())
			Expect(rerun.Body.Close()).To(Succeed())
			Expect(rerun.StatusCode).To(Equal(http.StatusAccepted))
		})

		Expect(held).To(BeTrue())
		Expect(reran).To(BeTrue())
	})
})
