package main

import (
	"time"

	"github.com/flanksource/gavel/fixtures"
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
})
