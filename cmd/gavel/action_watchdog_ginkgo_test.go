//go:build !windows

package main

import (
	"bytes"
	"io"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("action watchdog", func() {
	It("preserves a completed child exit code", func() {
		code := runActionWatchdog(actionWatchdogOptions{
			Executable: "sh",
			Args:       []string{"-c", "exit 7"},
			Timeout:    time.Second,
			Stdout:     io.Discard,
			Stderr:     io.Discard,
		})

		Expect(code).To(Equal(7))
	})

	It("requests a stack, kills a hung child tree, and returns timeout", func() {
		var output bytes.Buffer
		var stackRequests atomic.Int32
		started := time.Now()
		code := runActionWatchdog(actionWatchdogOptions{
			Executable: "sh",
			Args:       []string{"-c", "sleep 30 & wait"},
			Timeout:    50 * time.Millisecond,
			StackGrace: 20 * time.Millisecond,
			Stdout:     io.Discard,
			Stderr:     &output,
			RequestStack: func(int) error {
				stackRequests.Add(1)
				return nil
			},
		})

		Expect(code).To(Equal(actionTimeoutExitCode))
		Expect(stackRequests.Load()).To(Equal(int32(1)))
		Expect(time.Since(started)).To(BeNumerically("<", 2*time.Second))
		Expect(output.String()).To(ContainSubstring("exceeded action timeout"))
	})
})
