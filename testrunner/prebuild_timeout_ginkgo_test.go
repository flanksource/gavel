//go:build !windows

package testrunner

import (
	"context"
	"time"

	clickyexec "github.com/flanksource/clicky/exec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("pre-build process", func() {
	It("returns promptly and kills the process tree when the run context expires", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 75*time.Millisecond)
		DeferCleanup(cancel)
		process := clickyexec.NewExec("sh", "-c", "sleep 30 & wait").WithProcessGroup()

		started := time.Now()
		_, err := runPreBuildProcess(ctx, process, "Pre-build test")

		Expect(err).To(MatchError(context.DeadlineExceeded))
		Expect(time.Since(started)).To(BeNumerically("<", 2*time.Second))
		Eventually(process.IsRunning, time.Second).Should(BeFalse())
	})
})
