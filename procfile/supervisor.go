package procfile

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/flanksource/clicky"
	cexec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/verify"
)

// Options configures a Supervisor.
type Options struct {
	// Root is the project root; the Procfile, .env and .gavel/proc live here.
	// Defaults to the current working directory.
	Root string
	// Procfile overrides discovery (see Find).
	Procfile string
	// Names selects a subset of processes to register + auto-start; empty runs
	// the default set for the active profile.
	Names []string
	// Profile is the active profile; entries with `profiles` auto-start only when
	// it matches. Empty means the default profile.
	Profile string
	// Foreground multiplexes process output to stdout (the `proc run` form).
	Foreground bool
	// Config carries global defaults from .gavel.yaml.
	Config verify.ProcfileConfig
}

// managed is the supervisor's per-process bookkeeping. The supervised lifecycle
// (run/restart, ports, status, resources) is owned by the clicky
// SupervisedProcess; this struct only holds gavel-side wiring.
type managed struct {
	entry     Entry
	logPath   string
	colorIdx  int
	overlay   map[string]string
	opts      cexec.SuperviseOptions
	autostart bool

	mu      sync.Mutex
	proc    *cexec.SupervisedProcess
	logFD   *os.File
	running bool
}

// Supervisor runs and watches the processes from a Procfile. It is the sole
// writer of .gavel/proc/state.json; CLI commands read that file or talk to the
// control socket for live operations. Per-process supervision is delegated to
// clicky's SupervisedProcess.
type Supervisor struct {
	root       string
	dir        string
	procfile   string
	socket     string
	profile    string
	foreground bool
	width      int
	started    *time.Time

	procs  []*managed
	byName map[string]*managed

	mu           sync.Mutex // guards active, stopping, fullyStarted, listener
	active       int
	stopping     bool
	fullyStarted bool
	listener     net.Listener

	done         chan struct{} // closed when shutting down (signal or all-exited)
	persistMu    sync.Mutex    // serialises state.json writes
	stdoutMu     sync.Mutex    // serialises foreground multiplexed output
	shutdownOnce sync.Once
}

// NewSupervisor resolves the environment and per-process options from opts into
// a ready-to-Run supervisor. opts.Root and opts.Procfile must already be
// resolved (the manager owns discovery via resolveTarget).
func NewSupervisor(opts Options) (*Supervisor, error) {
	if opts.Root == "" || opts.Procfile == "" {
		return nil, fmt.Errorf("supervisor requires a resolved Root and Procfile")
	}
	root, _ := filepath.Abs(opts.Root)

	entries, err := Load(opts.Procfile)
	if err != nil {
		return nil, err
	}
	entries, err = Select(entries, opts.Names)
	if err != nil {
		return nil, err
	}
	dotenv, err := LoadDotEnv(filepath.Join(root, ".env"))
	if err != nil {
		return nil, err
	}
	// Capture the developer's login-shell environment (real PATH, etc.) once and
	// layer it under .env/config below, so supervised processes resolve commands
	// like npm/node even when the supervisor was started by a launchd/systemd
	// service that stripped the interactive shell env. Best-effort: a shell that
	// errors or times out must not wedge startup, so log loudly and fall back to
	// inheriting only the supervisor's own environment.
	userEnv, err := LoadUserEnv(root)
	if err != nil {
		logger.Warnf("load login-shell env for %s: %v", root, err)
	}
	dir, err := StateDir(root)
	if err != nil {
		return nil, err
	}

	profile := resolveProfile(opts)
	s := &Supervisor{root: root, dir: dir, procfile: opts.Procfile, profile: profile, foreground: opts.Foreground, byName: map[string]*managed{}, done: make(chan struct{})}
	for i, e := range entries {
		policy, maxR := resolvePolicy(opts.Config, e)
		limits, err := resolveLimits(opts.Config, e)
		if err != nil {
			return nil, err
		}
		// An explicitly named subset auto-starts entirely; otherwise an entry
		// auto-starts only when it is in the active profile and in the default set.
		inProfile := len(e.Profiles) == 0 || contains(e.Profiles, profile)
		autostart := len(opts.Names) > 0 || (inProfile && (e.Default == nil || *e.Default))
		m := &managed{
			entry:     e,
			logPath:   LogPath(dir, e.Name),
			colorIdx:  i,
			overlay:   MergeEnv(userEnv, dotenv, opts.Config.Env, e.Env),
			opts:      cexec.SuperviseOptions{Limits: limits, RestartPolicy: policy, MaxRestarts: maxR},
			autostart: autostart,
		}
		if len(e.Name) > s.width {
			s.width = len(e.Name)
		}
		s.procs = append(s.procs, m)
		s.byName[e.Name] = m
	}
	return s, nil
}

// Run starts the auto-start processes and blocks until the supervisor is stopped
// — either by a shutdown signal or because every running process exited with
// nothing left to restart — then tears everything down and returns.
func (s *Supervisor) Run() error {
	if err := s.Start(); err != nil {
		return err
	}
	s.Wait()
	return nil
}

// Start writes the supervisor pidfile, opens the control socket, builds every
// process's supervised unit, and launches the auto-start ones. Non-autostart
// units stay registered (stopped) so they can be started on demand. Returns once
// everything is launched (non-blocking).
func (s *Supervisor) Start() error {
	if err := WritePid(SupervisorPidPath(s.dir), os.Getpid()); err != nil {
		return err
	}
	now := time.Now()
	s.started = &now
	for _, m := range s.procs {
		if err := truncateFile(m.logPath); err != nil {
			return err
		}
		if err := s.build(m); err != nil {
			return err
		}
	}
	if err := s.serveControl(); err != nil {
		return err
	}

	for _, m := range s.procs {
		if m.autostart {
			s.startProc(m)
		}
	}
	s.persist()

	s.mu.Lock()
	s.fullyStarted = true
	allDone := s.active <= 0 && !s.stopping
	s.mu.Unlock()
	if allDone {
		s.beginShutdown()
	}
	return nil
}

// Wait blocks until a shutdown signal arrives or every process has exited, then
// tears the daemon down. A signal is an explicit stop and clears the state; every
// process exiting on its own keeps a terminal state.json so `proc status` and
// `proc start`'s readiness can see whether a process crashed.
func (s *Supervisor) Wait() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(sig)
	select {
	case <-sig:
		s.teardown(false)
	case <-s.done:
		s.teardown(true)
	}
}

// build opens the per-process log and constructs its clicky SupervisedProcess.
func (s *Supervisor) build(m *managed) error {
	logFD, err := os.OpenFile(m.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log %s: %w", m.logPath, err)
	}
	m.logFD = logFD

	var out io.Writer = logFD
	if s.foreground {
		out = io.MultiWriter(logFD, newPrefixWriter(os.Stdout, &s.stdoutMu, m.entry.Name, s.width, m.colorIdx))
	}

	name := m.entry.Name
	opts := m.opts
	opts.DetectPorts = true
	opts.OnStart = func() {
		fmt.Fprintf(logFD, "--- %s start %s ---\n", name, time.Now().Format(time.RFC3339))
	}
	opts.OnExit = func() { s.onUnitExit(m) }

	// The command is run as-is; env vars are injected via WithEnv and expanded by
	// the shell at runtime. Pre-expanding here would clobber shell-internal
	// variables like $i or arithmetic like $((i+1)).
	m.proc = clicky.Exec(m.entry.Command).
		WithCwd(s.root).
		WithEnv(m.overlay).
		WithProcessGroup().
		Stream(out, out).
		Supervise(opts)
	return nil
}

func (s *Supervisor) startProc(m *managed) {
	if s.isStopping() {
		return
	}
	m.mu.Lock()
	if m.running || m.proc == nil {
		m.mu.Unlock()
		return
	}
	m.running = true
	p := m.proc
	m.mu.Unlock()

	s.mu.Lock()
	s.active++
	s.mu.Unlock()
	p.Start()
}

func (s *Supervisor) stopProc(m *managed) {
	m.mu.Lock()
	p := m.proc
	m.mu.Unlock()
	if p != nil {
		p.Stop() // blocks until the loop ends → onUnitExit decrements active
	}
}

func (s *Supervisor) restartProc(m *managed) {
	m.mu.Lock()
	p := m.proc
	running := m.running
	m.mu.Unlock()
	if p == nil {
		return
	}
	if !running {
		s.startProc(m)
		return
	}
	p.Restart()
}

// onUnitExit is the OnExit callback for a unit: its supervise loop ended
// permanently (exited/crashed/stopped). Refresh state and decrement the live
// count, shutting the daemon down once nothing is left running.
func (s *Supervisor) onUnitExit(m *managed) {
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()
	s.persist()
	s.decActive()
}

// Shutdown stops every process (graceful then forceful), closes the control
// socket, and removes the state/pid files. It is idempotent and does not exit
// the process, so it is safe to call from tests. It is the explicit-stop path:
// the state is cleared.
func (s *Supervisor) Shutdown() {
	s.teardown(false)
}

// teardown stops every process (graceful then forceful), closes the control
// socket, and releases the live-supervisor artifacts. With persistTerminal it
// writes a final state.json capturing each process's terminal status/exit code
// (a post-mortem for a daemon that exited on its own) and keeps it; otherwise it
// clears the state entirely (an explicit stop). It is idempotent.
func (s *Supervisor) teardown(persistTerminal bool) {
	s.mu.Lock()
	if s.stopping {
		s.mu.Unlock()
		return
	}
	s.stopping = true
	l := s.listener
	s.mu.Unlock()

	s.beginShutdown() // wake any process in restart backoff
	if l != nil {
		_ = l.Close()
	}
	_ = os.Remove(ControlSocketPath(s.root))

	var wg sync.WaitGroup
	for _, m := range s.procs {
		m.mu.Lock()
		p := m.proc
		m.mu.Unlock()
		if p == nil {
			continue
		}
		wg.Add(1)
		go func(p *cexec.SupervisedProcess) {
			defer wg.Done()
			p.Stop()
		}(p)
	}
	wg.Wait()

	// Snapshot before the FDs close: status/exit codes are read off the live
	// supervised units. SupervisorPID is zeroed so the persisted state never
	// reports a (dead) supervisor as running.
	var terminal State
	if persistTerminal {
		terminal = s.State()
		terminal.SupervisorPID = 0
	}

	for _, m := range s.procs {
		m.mu.Lock()
		fd := m.logFD
		m.mu.Unlock()
		if fd != nil {
			_ = fd.Close()
		}
	}

	if persistTerminal {
		if err := WriteState(s.dir, terminal); err != nil {
			logger.Warnf("persist terminal proc state: %v", err)
		}
		_ = CleanPidfiles(s.dir)
	} else {
		_ = Clean(s.dir)
	}
}

func (s *Supervisor) decActive() {
	s.mu.Lock()
	s.active--
	trigger := s.active <= 0 && s.fullyStarted && !s.stopping
	s.mu.Unlock()
	if trigger {
		s.beginShutdown()
	}
}

func (s *Supervisor) beginShutdown() {
	s.shutdownOnce.Do(func() {
		close(s.done)
	})
}

func (s *Supervisor) isStopping() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopping
}

// persist writes the current state of every process to state.json. Writes are
// serialised and skipped once shutdown has begun so a late write can't recreate
// the file Clean just removed. Live status/ports/metrics are read directly off
// the supervised units via the control socket; state.json is the fallback for
// when the daemon isn't reachable.
func (s *Supervisor) persist() {
	if s.isStopping() {
		return
	}
	s.persistMu.Lock()
	defer s.persistMu.Unlock()
	if err := WriteState(s.dir, s.State()); err != nil {
		logger.Warnf("write proc state: %v", err)
	}
}

func (m *managed) snapshot() ProcState {
	m.mu.Lock()
	p := m.proc
	m.mu.Unlock()

	st := ProcState{
		Name:    m.entry.Name,
		Command: m.entry.Command,
		LogFile: m.logPath,
		Status:  StatusStopped,
	}
	if p == nil {
		return st
	}
	res := p.Resources()
	st.PID = p.Pid()
	st.Status = string(p.Status())
	st.Started = p.Started()
	st.Restarts = p.Restarts()
	st.ExitCode = p.ExitCode()
	st.Ports = p.Ports()
	st.CPUPercent = res.CPUPercent
	st.MemoryRSS = res.RSSBytes
	st.OpenFiles = res.OpenFiles
	st.Tree = procTree(p.Tree())
	return st
}

// procTree maps clicky's per-process samples to the wire ProcNode shape.
func procTree(samples []cexec.ProcessSample) []ProcNode {
	if len(samples) == 0 {
		return nil
	}
	nodes := make([]ProcNode, len(samples))
	for i, s := range samples {
		nodes[i] = ProcNode{
			PID:        int(s.PID),
			PPID:       int(s.PPID),
			Command:    s.Command,
			CPUPercent: s.CPUPercent,
			MemoryRSS:  s.RSSBytes,
			OpenFiles:  s.OpenFiles,
		}
	}
	return nodes
}

// resolveProfile picks the active profile: the --profile flag wins, else the
// .gavel.yaml default.
func resolveProfile(opts Options) string {
	if opts.Profile != "" {
		return opts.Profile
	}
	return opts.Config.Profile
}

// resolvePolicy resolves the restart policy + cap for an entry, falling back to
// the global config. Maps the config's verify.RestartPolicy to clicky's enum
// (identical values).
func resolvePolicy(cfg verify.ProcfileConfig, e Entry) (cexec.RestartPolicy, int) {
	policy := cfg.AutoRestart
	if e.AutoRestart != "" {
		policy = e.AutoRestart
	}
	maxR := cfg.MaxRestarts
	if e.MaxRestarts != nil {
		maxR = *e.MaxRestarts
	}
	return cexec.RestartPolicy(policy), maxR
}

// resolveLimits resolves the resource limits for an entry, falling back to the
// global config. An unparseable mem is a loud error so a typo fails the start.
func resolveLimits(cfg verify.ProcfileConfig, e Entry) (cexec.ResourceLimits, error) {
	mem := cfg.Mem
	if e.Mem != "" {
		mem = e.Mem
	}
	cpu := cfg.CPU
	if e.CPU != 0 {
		cpu = e.CPU
	}
	var rss uint64
	if mem != "" {
		parsed, err := humanize.ParseBytes(mem)
		if err != nil {
			return cexec.ResourceLimits{}, fmt.Errorf("invalid mem %q for process %q: %w", mem, e.Name, err)
		}
		rss = parsed
	}
	return cexec.ResourceLimits{MaxRSSBytes: rss, MaxCPUPercent: cpu}, nil
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func truncateFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("truncate %s: %w", path, err)
	}
	return f.Close()
}
