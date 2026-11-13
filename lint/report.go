package lint

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/git/changeset"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// ReportOptions is how ReportView renders a run: Summary collapses it to
// linter -> rule with the first SummaryLimit locations per rule (5 when unset).
type ReportOptions struct {
	Summary      bool
	SummaryLimit int
}

// ReportView is the value `gavel lint` hands its pretty formatter for results,
// so an embedder renders a lint run exactly as the CLI does. (It is not named
// Report: ginkgo's dot-import exports that name into this package's specs.)
func ReportView(results []*linters.LinterResult, opts ReportOptions) any {
	if opts.Summary {
		return linters.NewSummaryView(results, opts.SummaryLimit)
	}
	return results
}

// SnapshotOptions is one finished lint run: the directory it linted, the
// version of the tool that ran it, when it started, the files and change
// selection it was narrowed to and its results.
type SnapshotOptions struct {
	WorkDir string
	Version string
	Started time.Time
	// Files are the files or folders the run was limited to, relative to
	// WorkDir unless absolute; empty for a run over the whole tree.
	Files   []string
	Changes []*changeset.ChangeSet
	Results []*linters.LinterResult
}

// Snapshot is a finished lint run as `gavel ui serve` and `--failed` read it.
func Snapshot(opts SnapshotOptions) *testui.Snapshot {
	return &testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{
			Version:   opts.Version,
			Started:   opts.Started,
			Ended:     time.Now().UTC(),
			Kind:      "lint",
			LintFiles: absoluteFiles(opts.WorkDir, opts.Files),
			Changes:   opts.Changes,
		},
		Git:    snapshots.GitInfo(opts.WorkDir),
		Status: testui.SnapshotStatus{LintRun: true},
		Lint:   opts.Results,
	}
}

func absoluteFiles(workDir string, files []string) []string {
	if len(files) == 0 {
		return nil
	}
	abs := make([]string, len(files))
	for i, file := range files {
		if !filepath.IsAbs(file) {
			file = filepath.Join(workDir, file)
		}
		abs[i] = filepath.Clean(file)
	}
	return abs
}

// SaveSnapshot persists a lint snapshot under stateRoot's .gavel: keyed by
// HEAD (last.json and the branch pointer) when it has one, and always as its
// own run-*.json, whose path it returns.
func SaveSnapshot(stateRoot string, snapshot *testui.Snapshot) (string, error) {
	if snapshot.Git != nil && snapshot.Git.SHA != "" {
		path, err := snapshots.Save(stateRoot, snapshot)
		if err != nil {
			return "", fmt.Errorf("persist lint snapshot: %w", err)
		}
		logger.V(1).Infof("wrote snapshot to %s", path)
	}
	path, err := snapshots.SavePerRun(stateRoot, snapshot, snapshot.Metadata.Started, "")
	if err != nil {
		return "", fmt.Errorf("persist per-run lint snapshot: %w", err)
	}
	logger.V(1).Infof("wrote per-run snapshot to %s", path)
	return path, nil
}

// ServeOptions is the UI Serve shows a finished lint run on: Server when the
// caller already serves one (it streamed the run), else a new one on Addr,
// announced to Writer and opened in a browser when Open.
type ServeOptions struct {
	Server  *testui.Server
	Addr    string
	Writer  io.Writer
	Open    bool
	Results []*linters.LinterResult
}

// Serve shows results in the gavel UI until SIGINT, SIGTERM or ctx ends, and
// returns the run as the UI last held it: a rerun from the UI replaces it.
func Serve(ctx context.Context, opts ServeOptions) (testui.Snapshot, error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)
	srv := opts.Server
	if srv == nil {
		var err error
		if srv, _, _, err = testui.Listen(testui.ListenOptions{Addr: opts.Addr, Title: "Lint UI", Writer: opts.Writer, Open: opts.Open}); err != nil {
			return testui.Snapshot{}, err
		}
		srv.BeginRun("lint")
	}
	srv.SetLintResults(opts.Results)
	srv.MarkDone()
	select {
	case <-sig:
	case <-ctx.Done():
	}
	return srv.Snapshot(), nil
}
