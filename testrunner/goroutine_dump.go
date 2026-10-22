package testrunner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	clickyprocess "github.com/flanksource/clicky/process"
	"github.com/flanksource/commons-db/tracing/process"
	"github.com/flanksource/commons-db/tracing/stacks"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/testrunner/parsers"
)

const (
	// goroutineDumpWait bounds how long a SIGQUIT Go test binary gets to print
	// its goroutines and exit before its tree is killed.
	goroutineDumpWait = 3 * time.Second
	// goroutineDumpFlush lets `go test` relay the output of a binary that
	// exited before the tree is killed.
	goroutineDumpFlush = 250 * time.Millisecond
)

// goTestBinary is a Go test binary (`<pkg>.test`) running under a go test
// process: the process a goroutine dump is asked of. Dir is its working
// directory, the package's source directory, which names its package.
type goTestBinary struct {
	PID int
	// Key is its node's key in the run's process_tree.
	Key    string
	Dir    string
	QuitAt time.Time
	// Err is why the binary could not be asked for its dump, or its package
	// directory read.
	Err error
}

// isGoTestBinary reports whether a ps command line runs a `go test` binary.
func isGoTestBinary(command string) bool {
	fields := strings.Fields(command)
	return len(fields) > 0 && strings.HasSuffix(filepath.Base(fields[0]), ".test")
}

// goTestBinaries lists the Go test binaries running under rootPID.
func goTestBinaries(ctx context.Context, rootPID int) ([]goTestBinary, error) {
	snapshot, err := clickyprocess.Discover(ctx, clickyprocess.SnapshotOptions{})
	if err != nil {
		return nil, fmt.Errorf("list the go test binaries under pid %d: %w", rootPID, err)
	}
	var pids []int
	for _, p := range snapshot.Subtree(rootPID) {
		if p.PID != rootPID && isGoTestBinary(p.Command) {
			pids = append(pids, p.PID)
		}
	}
	snapshot.PopulateWorkingDirectories(ctx, pids)
	binaries := make([]goTestBinary, 0, len(pids))
	for _, pid := range pids {
		p, _ := snapshot.Get(pid)
		binary := goTestBinary{PID: pid, Key: process.Key(p), Dir: p.CWD}
		if p.CWDError != "" {
			binary.Err = fmt.Errorf("read the package directory of go test binary %d: %s", pid, p.CWDError)
		}
		binaries = append(binaries, binary)
	}
	return binaries, nil
}

// quitGoTestBinaries SIGQUITs every Go test binary under rootPID, so each
// prints its goroutines, and waits up to wait for them to exit and the
// output to be relayed. A failure to list or signal is logged and recorded on
// the binary; the caller kills the tree regardless.
func quitGoTestBinaries(rootPID int, wait time.Duration) []goTestBinary {
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	binaries, err := goTestBinaries(ctx, rootPID)
	if err != nil {
		logger.Warnf("goroutine dump: %v", err)
		return nil
	}
	for i := range binaries {
		binaries[i].QuitAt = time.Now().UTC()
		if err := signalQuit(binaries[i].PID); err != nil {
			binaries[i].Err = err
			logger.Warnf("goroutine dump: %v", err)
		}
	}
	waitForExit(binaries, wait)
	return binaries
}

// waitForExit waits until every binary exited, or wait passed, then for go
// test to relay what they printed.
func waitForExit(binaries []goTestBinary, wait time.Duration) {
	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		alive := false
		for _, binary := range binaries {
			alive = alive || processAlive(binary.PID)
		}
		if !alive {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(goroutineDumpFlush)
}

func processAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	return err == nil && signalAlive(process)
}

// checkpointTimeoutDump records, as a stacks checkpoint of the run's process
// trace on its Go test binary's key, the goroutine dump the timed-out package
// pkgPath printed, as the attempts carry it. Without a process trace the
// attempts are the record. A dump that cannot be recorded is logged: the
// run's verdict stands.
func (o *TestOrchestrator) checkpointTimeoutDump(pkgPath string, suites parsers.TestSuiteResults, dumped []goTestBinary) {
	if o.ProcessTrace == nil {
		return
	}
	binary, ok := binaryOf(o.WorkDir, pkgPath, dumped)
	if !ok {
		logger.Warnf("record the goroutine dump of %s: no go test binary of it was running when it timed out", pkgPath)
		return
	}
	text := timedOutStack(suites)
	switch {
	case binary.Err != nil:
		logger.Warnf("record the goroutine dump of %s: %v", pkgPath, binary.Err)
		return
	case text == "":
		logger.Warnf("record the goroutine dump of %s: go test binary %d printed no goroutine dump within %s of SIGQUIT", pkgPath, binary.PID, goroutineDumpWait)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := o.ProcessTrace.CheckpointStacks(ctx, stacks.CheckpointParams{
		Key: binary.Key, Reason: PackageTimeoutReason(pkgPath),
		Dump: &stacks.Dump{Format: stacks.FormatGo, Text: text, CollectedAt: binary.QuitAt},
	})
	if err != nil {
		logger.Warnf("record the goroutine dump of %s (pid %d) on the process trace: %v", pkgPath, binary.PID, err)
	}
}

// binaryOf is the dumped Go test binary of package pkgPath: the one running
// in its directory.
func binaryOf(workDir, pkgPath string, dumped []goTestBinary) (goTestBinary, bool) {
	dir := pkgPath
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(workDir, pkgPath)
	}
	for _, binary := range dumped {
		if sameDir(binary.Dir, dir) {
			return binary, true
		}
	}
	return goTestBinary{}, false
}

// goroutineCount is how many goroutines a Go dump holds: the records a stacks
// checkpoint of it stores, less a failed one.
func goroutineCount(dump string) int {
	if dump == "" {
		return 0
	}
	count := 0
	for _, record := range stacks.Records(stacks.Stack{}, stacks.Dump{Format: stacks.FormatGo, Text: dump}) {
		if record.Error == "" {
			count++
		}
	}
	return count
}

// timedOutStack is the first goroutine dump a timed-out test's attempt holds.
func timedOutStack(suites parsers.TestSuiteResults) string {
	var find func(tests parsers.Tests) string
	find = func(tests parsers.Tests) string {
		for _, test := range tests {
			if n := len(test.Attempts); n > 0 && test.Attempts[n-1].StackTrace != "" {
				return test.Attempts[n-1].StackTrace
			}
			if stack := find(test.Children); stack != "" {
				return stack
			}
		}
		return ""
	}
	for _, suite := range suites {
		if stack := find(suite.Tests); stack != "" {
			return stack
		}
	}
	return ""
}

// sameDir reports whether two directories name the same place, through
// symlinks such as macOS's /var -> /private/var.
func sameDir(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	resolve := func(dir string) string {
		if resolved, err := filepath.EvalSymlinks(dir); err == nil {
			return resolved
		}
		return filepath.Clean(dir)
	}
	return resolve(a) == resolve(b)
}
