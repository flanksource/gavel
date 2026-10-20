package testrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/commons-db/tracing/stacks"
	gopsutil "github.com/shirou/gopsutil/v4/process"
)

// TransportSigquit names gavel's stack transport for the Go test binaries it
// supervises: SIGQUIT, on which the Go runtime prints every goroutine and
// exits.
const TransportSigquit = "sigquit"

// ErrNotGoTestBinary refuses a dump of a process that is not a Go test binary
// a running go test batch of this run started.
var ErrNotGoTestBinary = errors.New("not a go test binary of a running go test batch")

// maxAncestry bounds how far up the parents of a test binary its batch is
// looked for: go test runs its binaries as its own children.
const maxAncestry = 8

// SigquitDumper is the stacks transport that SIGQUITs a Go test binary a
// running go test batch of this run started, which kills it, and returns the
// goroutines it printed, taken from its batch's output. A host registers it
// (traces.Host.RegisterDumper) for the run's stacks kind.
type SigquitDumper struct {
	mu      sync.Mutex
	batches map[*exec.Process]*goTestOutput
	dumpMu  sync.Mutex
}

var _ stacks.Dumper = (*SigquitDumper)(nil)

func NewSigquitDumper() *SigquitDumper {
	return &SigquitDumper{batches: map[*exec.Process]*goTestOutput{}}
}

func (d *SigquitDumper) Transport() string { return TransportSigquit }
func (d *SigquitDumper) Destructive() bool { return true }

// Available is nil for a Go test binary of a running batch.
func (d *SigquitDumper) Available(pid int) error {
	_, err := d.batchOf(context.Background(), pid)
	return err
}

// watch tees the go test batch process's output, to take dumps from, until
// the returned stop. Call it before the process starts.
func (d *SigquitDumper) watch(batch *exec.Process) func() {
	output := &goTestOutput{}
	batch.Stream(output, io.Discard)
	d.mu.Lock()
	d.batches[batch] = output
	d.mu.Unlock()
	return func() {
		d.mu.Lock()
		delete(d.batches, batch)
		d.mu.Unlock()
	}
}

// Dump SIGQUITs the Go test binary pid, which prints its goroutines and
// exits, and returns the dump its batch's output held. One dump is taken at a
// time.
func (d *SigquitDumper) Dump(ctx context.Context, pid int) (stacks.Dump, error) {
	d.dumpMu.Lock()
	defer d.dumpMu.Unlock()
	output, err := d.batchOf(ctx, pid)
	if err != nil {
		return stacks.Dump{}, err
	}
	capture := output.beginCapture()
	defer output.endCapture()
	collectedAt := time.Now().UTC()
	if err := signalQuit(pid); err != nil {
		return stacks.Dump{}, err
	}
	select {
	case <-capture.done:
	case <-time.After(goroutineDumpWait):
	case <-ctx.Done():
		return stacks.Dump{}, fmt.Errorf("goroutine dump of pid %d: %w", pid, ctx.Err())
	}
	text := extractStackTrace(output.captured(), "")
	if text == "" {
		return stacks.Dump{}, fmt.Errorf("go test binary %d printed no goroutine dump within %s of SIGQUIT", pid, goroutineDumpWait)
	}
	return stacks.Dump{Format: stacks.FormatGo, Text: text, CollectedAt: collectedAt}, nil
}

// batchOf is the output of the running batch pid is a Go test binary of.
func (d *SigquitDumper) batchOf(ctx context.Context, pid int) (*goTestOutput, error) {
	d.mu.Lock()
	batches := make(map[int32]*goTestOutput, len(d.batches))
	for batch, output := range d.batches {
		if root := batch.Pid(); root > 0 {
			batches[int32(root)] = output
		}
	}
	d.mu.Unlock()
	if len(batches) == 0 {
		return nil, fmt.Errorf("pid %d: %w: none is running", pid, ErrNotGoTestBinary)
	}
	p, err := gopsutil.NewProcessWithContext(ctx, int32(pid))
	if err != nil {
		return nil, fmt.Errorf("pid %d: %w: %w", pid, ErrNotGoTestBinary, err)
	}
	command, err := p.CmdlineWithContext(ctx)
	if err != nil || !isGoTestBinary(command) {
		return nil, fmt.Errorf("pid %d: %w", pid, ErrNotGoTestBinary)
	}
	for range maxAncestry {
		ppid, err := p.PpidWithContext(ctx)
		if err != nil || ppid <= 1 {
			break
		}
		if output, ok := batches[ppid]; ok {
			return output, nil
		}
		if p, err = gopsutil.NewProcessWithContext(ctx, ppid); err != nil {
			break
		}
	}
	return nil, fmt.Errorf("pid %d: %w", pid, ErrNotGoTestBinary)
}

// goTestEvent is the part of a `go test -json` event a dump is taken from.
type goTestEvent struct {
	Action  string
	Package string
	Test    string
	Output  string
}

// goTestOutput is a go test batch's -json stdout, split into events while a
// dump is being captured from it.
type goTestOutput struct {
	mu      sync.Mutex
	partial []byte
	capture *dumpCapture
}

// dumpCapture collects the output of the package that printed a SIGQUIT
// dump, from that line until the package ends (done).
type dumpCapture struct {
	pkg  string
	text strings.Builder
	done chan struct{}
}

func (o *goTestOutput) Write(p []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	lines := bytes.Split(append(o.partial, p...), []byte("\n"))
	o.partial = slices.Clone(lines[len(lines)-1])
	if o.capture == nil {
		return len(p), nil
	}
	for _, line := range lines[:len(lines)-1] {
		var event goTestEvent
		// go test prints build failures and the like as plain lines between
		// its JSON events; they carry no dump.
		if json.Unmarshal(line, &event) == nil {
			o.capture.observe(event)
		}
	}
	return len(p), nil
}

func (o *goTestOutput) beginCapture() *dumpCapture {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.capture = &dumpCapture{done: make(chan struct{})}
	return o.capture
}

func (o *goTestOutput) captured() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.capture.text.String()
}

func (o *goTestOutput) endCapture() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.capture = nil
}

func (c *dumpCapture) observe(event goTestEvent) {
	select {
	case <-c.done:
		return
	default:
	}
	if c.pkg == "" {
		if event.Action != "output" || !strings.Contains(event.Output, "SIGQUIT: quit") {
			return
		}
		c.pkg = event.Package
	}
	switch {
	case event.Package != c.pkg:
	case event.Action == "output":
		c.text.WriteString(event.Output)
	case event.Test == "" && (event.Action == "fail" || event.Action == "pass" || event.Action == "skip"):
		close(c.done)
	}
}
