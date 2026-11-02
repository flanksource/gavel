package testrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// goPackageDeadlines holds each package of a go test batch to its own
// --test-timeout, as when every package ran in a process of its own. It reads
// the batch's go test -json stream: a package's clock starts at its "start"
// event, which go test writes as it runs the package's test binary, and stops
// at its pass, fail or skip. A package past its deadline has its binary
// SIGQUIT for a goroutine dump, then killed; go test reports it failed and
// runs the rest. A package with no binary to kill gets the whole batch
// stopped through killAll, so the deadline still holds.
type goPackageDeadlines struct {
	budget  time.Duration
	process *exec.Process
	killAll func()

	mu       sync.Mutex
	partial  []byte
	timers   map[string]*time.Timer
	timedOut map[string]bool
	dumped   []goTestBinary
	pending  sync.WaitGroup
}

func newGoPackageDeadlines(budget time.Duration, process *exec.Process, killAll func()) *goPackageDeadlines {
	return &goPackageDeadlines{budget: budget, process: process, killAll: killAll, timers: map[string]*time.Timer{}, timedOut: map[string]bool{}}
}

func (d *goPackageDeadlines) Write(p []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.partial = append(d.partial, p...)
	for {
		i := bytes.IndexByte(d.partial, '\n')
		if i < 0 {
			break
		}
		d.observe(d.partial[:i])
		d.partial = d.partial[i+1:]
	}
	return len(p), nil
}

func (d *goPackageDeadlines) observe(line []byte) {
	var event struct{ Action, Package, Test string }
	if json.Unmarshal(line, &event) != nil || event.Package == "" || event.Test != "" {
		return
	}
	switch event.Action {
	case "start":
		if _, running := d.timers[event.Package]; running {
			return
		}
		pkg := event.Package
		d.pending.Add(1)
		d.timers[pkg] = time.AfterFunc(d.budget, func() {
			defer d.pending.Done()
			d.expire(pkg)
		})
	case "pass", "fail", "skip":
		if timer, running := d.timers[event.Package]; running && timer.Stop() {
			d.pending.Done()
		}
	}
}

func (d *goPackageDeadlines) expire(pkg string) {
	d.mu.Lock()
	d.timedOut[pkg] = true
	d.mu.Unlock()
	logger.Warnf("test package %s exceeded deadline %s — capturing diagnostics before killing its test binary", pkg, d.budget)
	captureGlobalDiagnostics(PackageTimeoutReason(pkg))
	ctx, cancel := context.WithTimeout(context.Background(), goroutineDumpWait)
	all, err := goTestBinaries(ctx, d.process.Pid())
	cancel()
	var binaries []goTestBinary
	for _, binary := range all {
		if importPathOfDir(binary.Dir) == pkg {
			binaries = append(binaries, binary)
		}
	}
	if err != nil || len(binaries) == 0 {
		logger.Warnf("no go test binary of %s to kill (%v); stopping its whole batch", pkg, err)
		d.killAll()
		return
	}
	for i := range binaries {
		binaries[i].QuitAt = time.Now().UTC()
		if err := signalQuit(binaries[i].PID); err != nil {
			binaries[i].Err = err
		}
	}
	waitForExit(binaries, goroutineDumpWait)
	for _, binary := range binaries {
		// A binary that could not be SIGQUIT (Windows) was never waited on.
		if binary.Err != nil || processAlive(binary.PID) {
			if process, err := os.FindProcess(binary.PID); err == nil {
				_ = process.Kill()
			}
		}
	}
	d.mu.Lock()
	d.dumped = append(d.dumped, binaries...)
	d.mu.Unlock()
}

// stop disarms the deadlines of packages still running and waits for any
// expiry in flight, so the results read after it are final.
func (d *goPackageDeadlines) stop() {
	d.mu.Lock()
	for _, timer := range d.timers {
		if timer.Stop() {
			d.pending.Done()
		}
	}
	d.mu.Unlock()
	d.pending.Wait()
}

// TimedOut reports whether package importPath ran past its deadline; call it
// after stop.
func (d *goPackageDeadlines) TimedOut(importPath string) bool { return d.timedOut[importPath] }

// Dumped are the binaries of timed-out packages that were SIGQUIT; call it
// after stop.
func (d *goPackageDeadlines) Dumped() []goTestBinary { return d.dumped }

// importPathOfDir is the import path of the Go package in dir, from the
// nearest go.mod above it; "" when there is none.
func importPathOfDir(dir string) string {
	if dir == "" {
		return ""
	}
	for root := filepath.Clean(dir); ; root = filepath.Dir(root) {
		if module, ok := parsers.GoModuleName(root); ok {
			rel, err := filepath.Rel(root, dir)
			if err != nil {
				return ""
			}
			if rel == "." {
				return module
			}
			return module + "/" + filepath.ToSlash(rel)
		}
		if filepath.Dir(root) == root {
			return ""
		}
	}
}
