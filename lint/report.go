package lint

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/git/changeset"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// SnapshotOptions is one finished lint run: the directory it linted, the
// version of the tool that ran it, when it started, the change selection it
// was narrowed to and its results.
type SnapshotOptions struct {
	WorkDir string
	Version string
	Started time.Time
	Changes []*changeset.ChangeSet
	Results []*linters.LinterResult
}

// Snapshot is a finished lint run as `gavel ui serve` and `--failed` read it.
func Snapshot(opts SnapshotOptions) *testui.Snapshot {
	return &testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{
			Version: opts.Version,
			Started: opts.Started,
			Ended:   time.Now().UTC(),
			Kind:    "lint",
			Changes: opts.Changes,
		},
		Git:    snapshots.GitInfo(opts.WorkDir),
		Status: testui.SnapshotStatus{LintRun: true},
		Lint:   opts.Results,
	}
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
