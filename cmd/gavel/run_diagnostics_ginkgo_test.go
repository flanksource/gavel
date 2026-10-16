package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons-db/tracing"
	tracedprocess "github.com/flanksource/commons-db/tracing/process"
	"github.com/flanksource/commons-db/tracing/stacks"
	"github.com/flanksource/gavel/traces"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeDiagnosticsTrace struct {
	stacks []stacks.CheckpointParams
	trees  int
	nodes  []tracedprocess.Node
	err    error
}

func (f *fakeDiagnosticsTrace) CheckpointStacks(_ context.Context, params stacks.CheckpointParams) error {
	f.stacks = append(f.stacks, params)
	return f.err
}

func (f *fakeDiagnosticsTrace) CheckpointTree(context.Context) error {
	f.trees++
	return nil
}

func (f *fakeDiagnosticsTrace) Nodes(context.Context) ([]tracedprocess.Node, error) {
	return f.nodes, nil
}

var _ = Describe("run diagnostics", func() {
	It("prints goroutines, active tasks, and the process tree its trace's records hold", func() {
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
			ProbeTimeout: 100 * time.Millisecond,
		})
		reporter.recordOn(&fakeDiagnosticsTrace{nodes: []tracedprocess.Node{
			{Key: "101@1", PID: 101, Command: "gavel test", Status: "running", IsRoot: true},
			{Key: "202@2", ParentKey: "101@1", PID: 202, PPID: 101, Command: "go test ./pkg", Status: "exited"},
		}})

		reporter.Capture("timeout")

		text := output.String()
		Expect(text).To(ContainSubstring("Gavel diagnostics: timeout"))
		Expect(text).To(ContainSubstring("goroutine fake-stack"))
		Expect(text).To(ContainSubstring("github.com/acme/pkg TestWidget"))
		Expect(text).NotTo(ContainSubstring("completed task"))
		Expect(text).To(ContainSubstring("  - pid=202 ppid=101 status=exited command=go test ./pkg"), "the child indented under its parent")
	})

	quietOptions := func(output io.Writer) runDiagnosticsOptions {
		return runDiagnosticsOptions{
			Output:       output,
			CaptureStack: func(io.Writer) error { return nil },
			CaptureTasks: func() []task.TaskSnapshot { return nil },
			ProbeTimeout: 100 * time.Millisecond,
		}
	}

	It("dumps gavel's goroutines into the run's stacks at each capture, tagged with its reason, and re-observes the tree", func() {
		var output bytes.Buffer
		reporter := newRunDiagnosticsReporter(quietOptions(&output))
		trace := &fakeDiagnosticsTrace{}
		reporter.recordOn(trace)

		reasons := []string{"periodic", "package-timeout:./pkg", "timeout"}
		for _, reason := range reasons {
			reporter.Capture(reason)
		}

		var want []stacks.CheckpointParams
		for _, reason := range reasons {
			want = append(want, stacks.CheckpointParams{PID: os.Getpid(), Transport: stacks.TransportInProcess, Reason: reason})
		}
		Expect(trace.stacks).To(Equal(want))
		Expect(trace.trees).To(Equal(len(reasons)))
		Expect(output.String()).To(ContainSubstring("checkpointed gavel's goroutines into the run's stacks"))
	})

	It("lands a watchdog checkpoint in the run's stacks store", func() {
		host, err := traces.Open(GinkgoT().TempDir(), traces.RunID(time.Now()))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(host.Close(context.Background())).To(Succeed()) })
		trace, err := host.StartProcessTrace(traces.ProcessTraceOptions{RootPID: os.Getpid()})
		Expect(err).NotTo(HaveOccurred())
		var output bytes.Buffer
		reporter := newRunDiagnosticsReporter(quietOptions(&output))
		reporter.recordOn(trace)

		reporter.Capture("periodic")

		var dumped []stacks.Stack
		Expect(tracing.ReadRecords(context.Background(), host.Runtime(), trace.Refs().Stacks, func(record stacks.Stack) error {
			dumped = append(dumped, record)
			return nil
		})).To(Succeed())
		Expect(dumped).To(ContainElement(And(
			HaveField("Reason", "periodic"), HaveField("Transport", stacks.TransportInProcess), HaveField("PID", os.Getpid()),
			HaveField("Frames", ContainElement(HaveField("Func", ContainSubstring("runDiagnosticsReporter")))),
		)))
		Expect(output.String()).To(ContainSubstring(fmt.Sprintf("- pid=%d", os.Getpid())), "the log's tree is the trace's records")
		Expect(trace.Stop(context.Background())).To(Succeed())
	})

	It("still logs, and says why nothing was checkpointed, without a process trace", func() {
		var output bytes.Buffer
		newRunDiagnosticsReporter(quietOptions(&output)).Capture("timeout")
		Expect(output.String()).To(And(
			ContainSubstring("Gavel diagnostics: timeout"),
			ContainSubstring("not checkpointed: the run has no process trace (it needs --ui or --diagnostics)"),
		))
	})

	It("reports a checkpoint that failed without failing the capture", func() {
		var output bytes.Buffer
		reporter := newRunDiagnosticsReporter(quietOptions(&output))
		reporter.recordOn(&fakeDiagnosticsTrace{err: errors.New("store closed")})
		reporter.Capture("periodic")
		Expect(output.String()).To(ContainSubstring("checkpoint failed: store closed"))
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
			ProbeTimeout: 20 * time.Millisecond,
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
