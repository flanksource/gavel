package types

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/slots"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fixture command slots", func() {
	DescribeTable("excludes slot waits from declared step timeouts", func(terminal string) {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("XDG_CACHE_HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		limits := slots.Limits{Global: 1}
		lease, err := slots.Acquire(context.Background(), slots.Request{Limits: limits, WorkDir: dir, Command: "blocker"})
		Expect(err).NotTo(HaveOccurred())
		defer lease.Close()
		nodes := parseYAMLDocument(fmt.Sprintf("kind: Fixture\nsteps:\n  - kind: exec\n    exec: printf\n    args: [finished]\n    terminal: '%s'\n    timeout: 150ms\n", terminal))
		released := make(chan error, 1)
		time.AfterFunc(350*time.Millisecond, func() { released <- lease.Close() })
		started := time.Now()
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: dir, ProcessRuntime: processruntime.Options{Slots: limits}})
		Expect(err).NotTo(HaveOccurred())
		Expect(<-released).To(Succeed())
		Expect(time.Since(started)).To(BeNumerically(">=", 350*time.Millisecond))
		Expect(results).To(HaveLen(1))
		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error)
		Expect(results[0].Stdout).To(Equal("finished"))
	}, Entry("pipe", ""), Entry("PTY", "pty"))
})
