package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/testrunner/ui/queryapp"
	"github.com/flanksource/gavel/testrunner/ui/runs"
	"github.com/flanksource/gavel/workspace"
)

// UIServeOptions configures the `gavel ui serve` subcommand.
//
// Two modes:
//
//   - Standalone: user invokes `gavel ui serve snapshot.json`
//     to replay a previously captured run. Binds --port (or picks ephemeral).
//   - Detached child: spawned by `gavel test --ui --detach`, which passes
//     --listener-fd=3 and GAVEL_UI_LOCKFILE=<path>. The child adopts the
//     inherited socket so the port is guaranteed to match what the parent
//     already announced to the user.
//
// Note: AutoStop and IdleTimeout are registered imperatively in init() rather
// than via struct tags because clicky's flag binder does not recognize
// time.Duration as a field type.
type UIServeOptions struct {
	Port         int           `flag:"port" help:"Bind this port (0 = pick ephemeral). Ignored when --listener-fd is set." default:"0"`
	Addr         string        `flag:"addr" help:"Interface to bind. Defaults to 0.0.0.0 (all interfaces); set localhost to restrict to this machine." default:"0.0.0.0"`
	ListenerFD   int           `flag:"listener-fd" help:"Adopt an inherited socket FD from the parent (internal: set by gavel test --ui --detach)."`
	ResultsFiles []string      `json:"-" args:"true"`
	AutoStop     time.Duration `json:"-"`
	IdleTimeout  time.Duration `json:"-"`
	URLFile      string        `flag:"url-file" help:"Write the bound URL to this path atomically so shell scripts can read it back."`
	Dev          bool          `flag:"dev" help:"Serve the page from the testrunner/ui Vite dev server, with clicky-ui from the sibling checkout's source (requires a source checkout)"`
}

// uiServeDurations holds the duration flags attached imperatively to the
// `gavel ui serve` cobra command. They are read inside runUIServe because
// clicky does not populate them into the options struct.
var uiServeDurations struct {
	AutoStop    time.Duration
	IdleTimeout time.Duration
}

func (o UIServeOptions) Help() api.Textable {
	return clicky.Text(`Run a detached gavel UI server that replays a captured test run.

Standalone replay of a JSON snapshot:

  gavel ui serve run.json --auto-stop=10m
  gavel ui serve run-1.json run-2.json --auto-stop=10m

Without snapshot files it serves the most recent test run and lint run
recorded under .gavel/run-*.json, in the nearest .gavel directory at or above
the working directory within its git repository (else the git root). The History tab
lists every recorded run, and each one opens at /runs/<run-id>/.

The server prints its URL on the first line of stdout so a wrapping script can
` + "`head -n1`" + ` it, and exits when either --auto-stop (hard wall clock from
start) or --idle-timeout (reset on every HTTP request) fires — whichever comes
first. A zero duration disables the corresponding timer.`)
}

func init() {
	serveCmd := clicky.AddNamedCommandWithContext("serve", uiCmd, UIServeOptions{}, runUIServe)
	// clicky cannot bind time.Duration fields directly; register them
	// imperatively so the flags show up in `gavel ui serve --help`.
	serveCmd.Flags().DurationVar(&uiServeDurations.AutoStop, "auto-stop", 30*time.Minute,
		"Hard wall-clock deadline from process start. 0 disables.")
	serveCmd.Flags().DurationVar(&uiServeDurations.IdleTimeout, "idle-timeout", 5*time.Minute,
		"Exit after this long with no HTTP requests. 0 disables.")
}

func runUIServe(ctx context.Context, opts UIServeOptions) (any, error) {
	defer keepTasksOffTerminal()()
	// Pull duration values from the imperative flag shim when opts don't
	// already supply them. In-process tests can set opts.AutoStop/IdleTimeout
	// directly to avoid touching the global flag state.
	if opts.AutoStop == 0 {
		opts.AutoStop = uiServeDurations.AutoStop
	}
	if opts.IdleTimeout == 0 {
		opts.IdleTimeout = uiServeDurations.IdleTimeout
	}

	listener, err := openListener(opts)
	if err != nil {
		return nil, err
	}

	dirs, err := workspace.From(ctx)
	if err != nil {
		listener.Close() //nolint:errcheck
		return nil, err
	}
	if len(opts.ResultsFiles) == 0 {
		if opts.ResultsFiles, err = latestRunFiles(dirs.StateRoot); err != nil {
			listener.Close() //nolint:errcheck
			return nil, err
		}
	}
	srv := testui.NewServer()
	if err := queryapp.Serve(srv); err != nil {
		listener.Close() //nolint:errcheck
		return nil, err
	}
	if err := loadResults(srv, opts.ResultsFiles...); err != nil {
		listener.Close() //nolint:errcheck
		return nil, fmt.Errorf("load results %v: %w", opts.ResultsFiles, err)
	}
	if opts.Dev {
		proxy, err := devProxy(testRunnerDev, "")
		if err != nil {
			listener.Close() //nolint:errcheck
			return nil, err
		}
		srv.SetDevProxy(proxy)
	}

	addr := listener.Addr().(*net.TCPAddr)
	url, networkURL := testui.URLs(boundHost(addr, opts.Addr), addr.Port)

	if opts.URLFile != "" {
		if err := writeURLFile(opts.URLFile, url); err != nil {
			logger.Warnf("Failed to write --url-file=%s: %v", opts.URLFile, err)
		}
	}

	// Print the URL on the first line of stdout so a wrapping script can
	// `head -n1` it regardless of how many log lines come after.
	fmt.Println(testui.Banner("UI", url, networkURL))

	// Notify the fork parent that we're ready. On Unix with a lockfile set in
	// the environment, this writes our PID into the JSON lockfile and releases
	// the exclusive flock so the parent's blocking poll can observe the
	// handoff. On other platforms this is a no-op.
	if err := notifyHandoff(opts, addr.Port, url); err != nil {
		logger.Warnf("Handoff notify failed (parent may time out): %v", err)
	}

	history := runs.Handler(runs.Options{WorkDir: dirs.StateRoot, Latest: srv.Handler(), LatestRunIDs: runIDsOf(opts.ResultsFiles)})
	idleCh := make(chan struct{}, 1)
	go http.Serve(listener, withIdleTimer(history, idleCh)) //nolint:errcheck

	return nil, errors.Join(serveUntilTimeout(opts.AutoStop, opts.IdleTimeout, idleCh), srv.CloseTraces(), runs.Close(history))
}

// runIDsOf returns the RunIDs of the per-run snapshots among paths, so the
// history view can mark which runs the default view shows.
func runIDsOf(paths []string) []string {
	var ids []string
	for _, path := range paths {
		if name := strings.TrimSuffix(filepath.Base(path), ".json"); strings.HasPrefix(name, snapshots.PerRunPrefix) {
			ids = append(ids, name)
		}
	}
	return ids
}

// serveUntilTimeout blocks until either hardTimeout fires or idleTimeout has
// elapsed with no activity on idleCh. A zero duration disables the
// corresponding timer. Returning nil signals clean shutdown.
func serveUntilTimeout(hardTimeout, idleTimeout time.Duration, idleCh <-chan struct{}) error {
	var hard <-chan time.Time
	if hardTimeout > 0 {
		hard = time.After(hardTimeout)
	}

	var idle *time.Timer
	var idleC <-chan time.Time
	if idleTimeout > 0 {
		idle = time.NewTimer(idleTimeout)
		idleC = idle.C
		defer idle.Stop()
	}

	for {
		select {
		case <-hard:
			logger.V(1).Infof("auto-stop: hard deadline %s reached", hardTimeout)
			return nil
		case <-idleC:
			logger.V(1).Infof("auto-stop: idle %s elapsed", idleTimeout)
			return nil
		case <-idleCh:
			if idle != nil {
				if !idle.Stop() {
					select {
					case <-idle.C:
					default:
					}
				}
				idle.Reset(idleTimeout)
			}
		}
	}
}

// withIdleTimer wraps a handler so every request emits on idleCh. The
// listener for idleCh (serveUntilTimeout) resets the idle timer on receipt.
// A non-blocking send means bursts of requests are coalesced — good enough
// because any single event resets the timer.
func withIdleTimer(next http.Handler, idleCh chan<- struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case idleCh <- struct{}{}:
		default:
		}
		next.ServeHTTP(w, r)
	})
}

// latestRunFiles resolves the newest test and lint runs recorded under
// <cwd>/.gavel, used when `gavel ui serve` is given no snapshot files.
func latestRunFiles(stateRoot string) ([]string, error) {
	paths, err := snapshots.LatestRuns(stateRoot)
	if err != nil {
		return nil, fmt.Errorf("find previous runs in %s: %w", filepath.Join(stateRoot, snapshots.Dir), err)
	}
	if len(paths) == 0 {
		logger.Infof("No previous test or lint runs in %s", filepath.Join(stateRoot, snapshots.Dir))
	}
	for _, path := range paths {
		logger.Infof("Loading previous run %s", path)
	}
	return paths, nil
}

func loadResults(srv *testui.Server, paths ...string) error {
	if len(paths) == 0 {
		srv.MarkDone()
		return nil
	}

	var merged testui.Snapshot
	var runs []testui.TraceRun
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var payload testui.Snapshot
		var shape struct {
			Status json.RawMessage `json:"status"`
		}
		if err := json.Unmarshal(data, &shape); err != nil {
			return fmt.Errorf("parse result %s: %w", path, err)
		}
		if len(shape.Status) == 0 || shape.Status[0] != '"' && shape.Status[0] != '{' {
			return fmt.Errorf("unsupported result format in %s: status must be an object or a legacy benchmark string", path)
		}
		if len(shape.Status) > 0 && shape.Status[0] == '"' {
			var legacy fixtures.BenchmarkReport
			if err := json.Unmarshal(data, &legacy); err != nil {
				return fmt.Errorf("parse legacy benchmark %s: %w", path, err)
			}
			if legacy.Version != 1 || legacy.Mode != "benchmark" && legacy.Mode != "profile" {
				return fmt.Errorf("unsupported legacy benchmark format in %s: version=%d mode=%q", path, legacy.Version, legacy.Mode)
			}
			if legacy.StartedAt.IsZero() || legacy.FinishedAt.IsZero() || legacy.FinishedAt.Before(legacy.StartedAt) {
				return fmt.Errorf("invalid legacy benchmark timestamps in %s", path)
			}
			payload = testui.Snapshot{Status: testui.SnapshotStatus{Running: false},
				Metadata:    &testui.SnapshotMetadata{Version: "1", Started: legacy.StartedAt, Ended: legacy.FinishedAt, Kind: "fixtures"},
				Tests:       fixtures.BenchmarkReportToTests(&legacy),
				Performance: fixtures.BenchmarkPerformanceFromReport(&legacy, path)}
		} else if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("parse snapshot %s: %w", path, err)
		}
		run, err := testui.TraceRunOf(payload.Metadata)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if run != nil {
			runs = append(runs, *run)
		}
		merged = mergeSnapshots(merged, payload)
	}
	srv.LoadSnapshot(merged)
	return srv.ReopenTraces(runs...)
}

func mergeSnapshots(dst, src testui.Snapshot) testui.Snapshot {
	dst.Tests = append(dst.Tests, src.Tests...)
	dst.Lint = append(dst.Lint, src.Lint...)
	if src.Bench != nil {
		dst.Bench = src.Bench
	}
	if src.FixtureBenchmark != nil {
		dst.FixtureBenchmark = src.FixtureBenchmark
	}
	if src.Performance != nil {
		dst.Performance = src.Performance
	}
	if src.Error != "" {
		dst.Error = src.Error
	}
	if src.LogTail != "" {
		dst.LogTail = src.LogTail
	}
	dst.Status.Running = dst.Status.Running || src.Status.Running
	dst.Status.LintRun = dst.Status.LintRun || src.Status.LintRun
	dst.Metadata = mergeSnapshotMetadata(dst.Metadata, src.Metadata)
	dst.Git = mergeSnapshotGit(dst.Git, src.Git)
	return dst
}

func mergeSnapshotMetadata(dst, src *testui.SnapshotMetadata) *testui.SnapshotMetadata {
	if src == nil {
		return dst
	}
	if dst == nil {
		cloned := *src
		if src.Args != nil {
			cloned.Args = make(map[string]any, len(src.Args))
			for k, v := range src.Args {
				cloned.Args[k] = v
			}
		}
		return &cloned
	}
	if src.Version != "" {
		dst.Version = src.Version
	}
	if !src.Started.IsZero() && (dst.Started.IsZero() || src.Started.Before(dst.Started)) {
		dst.Started = src.Started
	}
	if !src.Ended.IsZero() && (dst.Ended.IsZero() || src.Ended.After(dst.Ended)) {
		dst.Ended = src.Ended
	}
	if src.Kind != "" {
		dst.Kind = src.Kind
	}
	if src.Sequence > dst.Sequence {
		dst.Sequence = src.Sequence
	}
	if src.RunID != "" {
		dst.RunID, dst.ProcessTrace = src.RunID, src.ProcessTrace
	}
	if src.Args != nil {
		dst.Args = make(map[string]any, len(src.Args))
		for k, v := range src.Args {
			dst.Args[k] = v
		}
	}
	return dst
}

func mergeSnapshotGit(dst, src *testui.SnapshotGit) *testui.SnapshotGit {
	if src == nil {
		return dst
	}
	if dst == nil {
		cloned := *src
		return &cloned
	}
	if src.Repo != "" {
		dst.Repo = src.Repo
	}
	if src.Root != "" {
		dst.Root = src.Root
	}
	if src.SHA != "" {
		dst.SHA = src.SHA
	}
	return dst
}

// boundHost returns the host a listener that is already open actually answers
// on. The socket is authoritative: whenever it is bound to a concrete address,
// that address is the only one it answers on, so --addr cannot override it.
//
// This matters for the inherited-socket path (`gavel test --ui --detach` passes
// an already-bound listener on fd 3): the parent chose the bind address, --addr
// keeps its 0.0.0.0 default, and announcing a private IP for a socket listening
// only on loopback gives a URL that is refused. Only a wildcard bind leaves the
// interface an open question, and that is where --addr decides.
func boundHost(addr *net.TCPAddr, requested string) string {
	if addr == nil || addr.IP == nil || addr.IP.IsUnspecified() {
		return requested
	}
	return addr.IP.String()
}

// writeURLFile writes url to path atomically (write to tempfile in same dir,
// fsync, rename). Atomicity matters because a wrapping shell script may be
// polling the file concurrently.
func writeURLFile(path, url string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".gavel-url-")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp) //nolint:errcheck
	if _, err := f.WriteString(url + "\n"); err != nil {
		f.Close() //nolint:errcheck
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close() //nolint:errcheck
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
