package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	testui "github.com/flanksource/gavel/testrunner/ui"
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
	Addr         string        `flag:"addr" help:"Interface to bind. Use 0.0.0.0 to expose on the LAN." default:"localhost"`
	ListenerFD   int           `flag:"listener-fd" help:"Adopt an inherited socket FD from the parent (internal: set by gavel test --ui --detach)."`
	ResultsFiles []string      `json:"-" args:"true"`
	AutoStop     time.Duration `json:"-"`
	IdleTimeout  time.Duration `json:"-"`
	URLFile      string        `flag:"url-file" help:"Write the bound URL to this path atomically so shell scripts can read it back."`
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

  gavel ui serve run.json --auto-stop=10m
  gavel ui serve run-1.json run-2.json --auto-stop=10m

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
	if len(opts.ResultsFiles) > 0 {
		if err := loadResults(srv, opts.ResultsFiles...); err != nil {
			listener.Close() //nolint:errcheck
			return nil, fmt.Errorf("load results %v: %w", opts.ResultsFiles, err)
		}
	} else {
		root, err := resolveGavelRoot()
		if err != nil {
			listener.Close() //nolint:errcheck
			return nil, err
		}
		srv.SetGavelDir(root)
		srv.MarkDone()
		logger.Infof("Indexing snapshots from %s", filepath.Join(root, ".gavel"))
	}

	addr := listener.Addr().(*net.TCPAddr)
	url := fmt.Sprintf("http://%s", net.JoinHostPort(announceHost(opts.Addr), strconv.Itoa(addr.Port)))

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

func loadResults(srv *testui.Server, paths ...string) error {
	if len(paths) == 0 {
		srv.MarkDone()
		return nil
	}

	var merged testui.Snapshot
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var payload testui.Snapshot
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("parse snapshot %s: %w", path, err)
		}
		merged = mergeSnapshots(merged, payload)
	}
	srv.LoadSnapshot(merged)
	return nil
}

func mergeSnapshots(dst, src testui.Snapshot) testui.Snapshot {
	dst.Tests = append(dst.Tests, src.Tests...)
	dst.Lint = append(dst.Lint, src.Lint...)
	if src.Bench != nil {
		dst.Bench = src.Bench
	}
	if src.Diagnostics != nil {
		dst.Diagnostics = src.Diagnostics
	}
	dst.Status.Running = dst.Status.Running || src.Status.Running
	dst.Status.LintRun = dst.Status.LintRun || src.Status.LintRun
	dst.Status.DiagnosticsAvailable = dst.Status.DiagnosticsAvailable || src.Status.DiagnosticsAvailable || src.Diagnostics != nil
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

// announceHost picks the hostname to print in the "UI at ..." banner given
// the user's --addr choice. When the user bound to a wildcard (0.0.0.0 or ::)
// it walks net.InterfaceAddrs() and returns the first non-loopback IPv4 so
// the printed URL is reachable from another host. Falls back to "localhost"
// if no external interface is found (e.g. sandboxed CI).
func announceHost(requested string) string {
	switch requested {
	case "", "localhost", "127.0.0.1", "::1":
		return "localhost"
	case "0.0.0.0", "::":
		if ip := firstNonLoopbackIPv4(); ip != "" {
			return ip
		}
		logger.Warnf("--addr=%s but no non-loopback IPv4 interface found; printing localhost", requested)
		return "localhost"
	default:
		return requested
	}
}

func firstNonLoopbackIPv4() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Warnf("net.InterfaceAddrs failed: %v", err)
		return ""
	}
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() || ipnet.IP.IsLinkLocalUnicast() {
			continue
		}
		v4 := ipnet.IP.To4()
		if v4 == nil {
			continue
		}
		return v4.String()
	}
	return ""
}

// resolveGavelRoot returns the git root of the current working directory and
// validates that .gavel/ exists inside it. Used by `gavel ui serve` (no args)
// to discover the snapshot directory to index.
func resolveGavelRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository: %w", err)
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", fmt.Errorf("git rev-parse --show-toplevel returned empty")
	}
	gavelDir := filepath.Join(root, ".gavel")
	if _, err := os.Stat(gavelDir); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no %s directory found in %s — run `gavel test` to populate it, or pass a snapshot path explicitly", ".gavel", root)
		}
		return "", err
	}
	return root, nil
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
