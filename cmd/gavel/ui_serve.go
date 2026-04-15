package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// UIServeOptions configures the `gavel ui serve` subcommand.
//
// Two modes:
//
//   - Standalone: user invokes `gavel ui serve --results-file=snapshot.json`
//     to replay a previously captured run. Binds --port (or picks ephemeral).
//   - Detached child: spawned by `gavel test --ui --auto-stop`, which passes
//     --listener-fd=3 and GAVEL_UI_LOCKFILE=<path>. The child adopts the
//     inherited socket so the port is guaranteed to match what the parent
//     already announced to the user.
//
// Note: AutoStop and IdleTimeout are registered imperatively in init() rather
// than via struct tags because clicky's flag binder does not recognize
// time.Duration as a field type.
type UIServeOptions struct {
	Port        int           `flag:"port" help:"Bind this port (0 = pick ephemeral). Ignored when --listener-fd is set." default:"0"`
	ListenerFD  int           `flag:"listener-fd" help:"Adopt an inherited socket FD from the parent (internal: set by gavel test --ui --auto-stop)."`
	ResultsFile string        `flag:"results-file" help:"Path to a JSON snapshot to load at startup. Written by the parent before fork."`
	AutoStop    time.Duration `json:"-"`
	IdleTimeout time.Duration `json:"-"`
	URLFile     string        `flag:"url-file" help:"Write the bound URL to this path atomically so shell scripts can read it back."`
}

// uiServeDurations holds the duration flags attached imperatively to the
// `gavel ui serve` cobra command. They are read inside runUIServe because
// clicky does not populate them into the options struct.
var uiServeDurations struct {
	AutoStop    time.Duration
	IdleTimeout time.Duration
}

func (o UIServeOptions) Help() string {
	return `Run a detached gavel UI server that replays a captured test run.

Standalone replay of a JSON snapshot:

  gavel ui serve --results-file=run.json --auto-stop=10m

The server prints its URL on the first line of stdout so a wrapping script can
` + "`head -n1`" + ` it, and exits when either --auto-stop (hard wall clock from
start) or --idle-timeout (reset on every HTTP request) fires — whichever comes
first. A zero duration disables the corresponding timer.`
}

func init() {
	serveCmd := clicky.AddNamedCommand("serve", uiCmd, UIServeOptions{}, runUIServe)
	// clicky cannot bind time.Duration fields directly; register them
	// imperatively so the flags show up in `gavel ui serve --help`.
	serveCmd.Flags().DurationVar(&uiServeDurations.AutoStop, "auto-stop", 30*time.Minute,
		"Hard wall-clock deadline from process start. 0 disables.")
	serveCmd.Flags().DurationVar(&uiServeDurations.IdleTimeout, "idle-timeout", 5*time.Minute,
		"Exit after this long with no HTTP requests. 0 disables.")
}

// snapshotPayload is the on-disk shape written by the fork parent and read by
// the detached child. Kept minimal so the parent doesn't have to import the
// full ui snapshot type.
type snapshotPayload struct {
	Tests []parsers.Test          `json:"tests"`
	Lint  []*linters.LinterResult `json:"lint,omitempty"`
}

func runUIServe(opts UIServeOptions) (any, error) {
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

	srv := testui.NewServer()
	if opts.ResultsFile != "" {
		if err := loadResults(srv, opts.ResultsFile); err != nil {
			listener.Close() //nolint:errcheck
			return nil, fmt.Errorf("load results %s: %w", opts.ResultsFile, err)
		}
	}
	srv.MarkDone()

	addr := listener.Addr().(*net.TCPAddr)
	url := fmt.Sprintf("http://localhost:%d", addr.Port)

	if opts.URLFile != "" {
		if err := writeURLFile(opts.URLFile, url); err != nil {
			logger.Warnf("Failed to write --url-file=%s: %v", opts.URLFile, err)
		}
	}

	// Print the URL on the first line of stdout so a wrapping script can
	// `head -n1` it regardless of how many log lines come after.
	fmt.Printf("UI at %s\n", url)

	// Notify the fork parent that we're ready. On Unix with a lockfile set in
	// the environment, this writes our PID into the JSON lockfile and releases
	// the exclusive flock so the parent's blocking poll can observe the
	// handoff. On other platforms this is a no-op.
	if err := notifyHandoff(opts, addr.Port, url); err != nil {
		logger.Warnf("Handoff notify failed (parent may time out): %v", err)
	}

	idleCh := make(chan struct{}, 1)
	handler := withIdleTimer(srv.Handler(), idleCh)
	go http.Serve(listener, handler) //nolint:errcheck

	return nil, serveUntilTimeout(opts.AutoStop, opts.IdleTimeout, idleCh)
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

func loadResults(srv *testui.Server, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var payload snapshotPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("parse snapshot: %w", err)
	}
	srv.SetResults(payload.Tests)
	if len(payload.Lint) > 0 {
		srv.SetLintResults(payload.Lint)
	}
	return nil
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
