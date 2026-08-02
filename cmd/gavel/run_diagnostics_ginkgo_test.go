package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/clicky/task"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("run diagnostics", func() {
	It("prints goroutines, active tasks, and the process tree", func() {
		var output bytes.Buffer
		reporter := newRunDiagnosticsReporter(runDiagnosticsOptions{
			Output: &output,
			CaptureStack: func(w io.Writer) error {
				_, err := io.WriteString(w, "goroutine fake-stack\n")
				return err
			},
			CaptureTasks: func() []task.TaskSnapshot {
				return []task.TaskSnapshot{
					{Name: "test group", Type: "group", Status: string(task.StatusRunning)},
					{Name: "github.com/acme/pkg TestWidget", Type: "task", Status: string(task.StatusRunning)},
					{Name: "completed task", Type: "task", Status: string(task.StatusSuccess)},
				}
			},
			CaptureProcesses: func() (*testui.DiagnosticsSnapshot, error) {
				return &testui.DiagnosticsSnapshot{Root: &testui.ProcessNode{
					PID:      101,
					Command:  "gavel test",
					Children: []*testui.ProcessNode{{PID: 202, Command: "go test ./pkg"}},
				}}, nil
			},
			ProbeTimeout: 100 * time.Millisecond,
		})

		reporter.Capture("timeout")

		text := output.String()
		Expect(text).To(ContainSubstring("Gavel diagnostics: timeout"))
		Expect(text).To(ContainSubstring("goroutine fake-stack"))
		Expect(text).To(ContainSubstring("github.com/acme/pkg TestWidget"))
		Expect(text).NotTo(ContainSubstring("completed task"))
		Expect(text).To(ContainSubstring("go test ./pkg"))
	})

	It("does not let a blocked task snapshot delay cancellation", func() {
		var output bytes.Buffer
		blocked := make(chan struct{})
		reporter := newRunDiagnosticsReporter(runDiagnosticsOptions{
			Output:       &output,
			CaptureStack: func(io.Writer) error { return nil },
			CaptureTasks: func() []task.TaskSnapshot {
				<-blocked
				return nil
			},
			CaptureProcesses: func() (*testui.DiagnosticsSnapshot, error) { return nil, nil },
			ProbeTimeout:     20 * time.Millisecond,
		})

		started := time.Now()
		reporter.Capture("timeout")
		close(blocked)

		Expect(time.Since(started)).To(BeNumerically("<", 200*time.Millisecond))
		Expect(output.String()).To(ContainSubstring("task snapshot timed out"))
	})

	It("reports periodically and once before cancelling at the deadline", func() {
		var mu sync.Mutex
		var reasons []string
		report := func(reason string) {
			mu.Lock()
			defer mu.Unlock()
			reasons = append(reasons, reason)
		}

		ctx, cancel := newMonitoredStopContext(context.Background(), monitoredStopOptions{
			Timeout:  55 * time.Millisecond,
			Interval: 20 * time.Millisecond,
			Report:   report,
		})
		DeferCleanup(cancel)

		Eventually(ctx.Done(), time.Second).Should(BeClosed())
		Expect(ctx.Err()).To(MatchError(context.DeadlineExceeded))

		mu.Lock()
		joined := strings.Join(reasons, ",")
		last := reasons[len(reasons)-1]
		mu.Unlock()
		Expect(joined).To(ContainSubstring("periodic"))
		Expect(last).To(Equal("timeout"))
	})
})
