package processruntime

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"github.com/flanksource/gavel/slots"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command slots", func() {
	It("releases both pools after a command timeout", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("XDG_CACHE_HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		git := exec.Command("git", "init", "-q", dir)
		Expect(git.Run()).To(Succeed())
		limits, timeout := slots.Limits{Global: 1, Project: 1}, "150ms"
		cmd := NewCommand(context.Background(), Options{Slots: limits, Timeout: &timeout}, "sleep", "10")
		cmd.Dir = dir
		Expect(cmd.Run()).To(HaveOccurred())
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		lease, err := slots.Acquire(ctx, slots.Request{Limits: limits, WorkDir: dir, Command: "after timeout"})
		Expect(err).NotTo(HaveOccurred())
		Expect(lease.Close()).To(Succeed())
	})
	It("waits before starting the command's own timeout", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("XDG_CACHE_HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		limits := slots.Limits{Global: 1}
		lease, err := slots.Acquire(context.Background(), slots.Request{Limits: limits, WorkDir: dir, Command: "blocker"})
		Expect(err).NotTo(HaveOccurred())
		defer lease.Close()
		timeout := "150ms"
		cmd := NewCommand(context.Background(), Options{Slots: limits, Timeout: &timeout}, "printf", "finished")
		cmd.Dir = dir
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		released := make(chan error, 1)
		time.AfterFunc(350*time.Millisecond, func() { released <- lease.Close() })
		started := time.Now()
		Expect(cmd.Run()).To(Succeed())
		Expect(<-released).To(Succeed())
		Expect(time.Since(started)).To(BeNumerically(">=", 350*time.Millisecond))
		Expect(stdout.String()).To(Equal("finished"))
	})
	It("preserves native exit errors with no slots configured", func() {
		err := NewCommand(context.Background(), Options{}, "sh", "-c", "exit 7").Run()
		Expect(err).To(BeAssignableToTypeOf(&exec.ExitError{}))
		Expect(err.(*exec.ExitError).ExitCode()).To(Equal(7))
	})
})
