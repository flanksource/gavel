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

	"github.com/flanksource/clicky"
	cexec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
)

// Restart policy values (see verify.ProcfileConfig.RestartPolicy).
const (
	RestartNo        = "no"
	RestartOnFailure = "on-failure"
	RestartAlways    = "always"
)

// stopGrace is how long a process is given to exit after SIGTERM before the
// supervisor escalates to SIGKILL of its whole process group.
const stopGrace = 5 * time.Second

// Port-detection cadence: after a process starts, poll for the TCP ports it is
// listening on every portPollInterval until they appear, the process exits, or
// portDetectTimeout elapses (a process that never binds, e.g. a worker).
const (
	portPollInterval  = 300 * time.Millisecond
	portDetectTimeout = 30 * time.Second
)

// Options configures a Supervisor.
type Options struct {
	// Root is the project root; the Procfile, .env and .gavel/proc live here.
	// Defaults to the current working directory.
	Root string
	// Procfile overrides discovery (see Find).
	Procfile string
	// Names selects a subset of processes; empty runs all.
	Names []string
	// Foreground multiplexes process output to stdout (the `proc run` form).
	Foreground bool
	// Config carries restart policy and env overrides from .gavel.yaml.
	Config verify.ProcfileConfig
}

// managed is the supervisor's per-process bookkeeping. Its mutex guards the
// fields a control goroutine may touch while runLoop is executing.
type managed struct {
	entry       Entry
	policy      string
	maxRestarts int
	overlay     map[string]string
	logPath     string
	colorIdx    int

	mu          sync.Mutex
	proc        *cexec.Process
	desired     bool
	active      bool
	status      string
	started     *time.Time
	restarts    int
	exitCode    *int
	ports       []int
	expectsPort bool // sticky: this process has listened on a port before, so a
	// (re)start is only "running" once a port is detected again.
	gen int
}

// Supervisor runs and watches the processes from a Procfile. It is the sole
// writer of .gavel/proc/state.json; CLI commands read that file or talk to the
// control socket for live operations.
type Supervisor struct {
	root       string
	dir        string
	procfile   string
	socket     string
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

	done         chan struct{}  // closed when shutting down (signal or all-exited)
	loops        sync.WaitGroup // tracks live runLoop goroutines
	persistMu    sync.Mutex     // serialises state.json writes
	stdoutMu     sync.Mutex     // serialises foreground multiplexed output
	shutdownOnce sync.Once
}

// NewSupervisor resolves the environment and per-process policy from opts into a
// ready-to-Run supervisor. opts.Root and opts.Procfile must already be resolved
// (the manager owns discovery via resolveTarget) so the state directory is
// stable across invocations from any subdirectory.
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

	s := &Supervisor{root: root, dir: dir, procfile: opts.Procfile, foreground: opts.Foreground, byName: map[string]*managed{}, done: make(chan struct{})}
	for i, e := range entries {
		policy, maxR := resolvePolicy(opts.Config, e.Name)
		m := &managed{
			entry:       e,
			policy:      policy,
			maxRestarts: maxR,
			overlay:     MergeEnv(userEnv, dotenv, opts.Config.Env, processEnv(opts.Config, e.Name)),
			logPath:     LogPath(dir, e.Name),
			colorIdx:    i,
			status:      StatusStopped,
		}
		if len(e.Name) > s.width {
			s.width = len(e.Name)
		}
		s.procs = append(s.procs, m)
		s.byName[e.Name] = m
	}
	return s, nil
}

// Run starts every process and blocks until the supervisor is stopped — either
// by a shutdown signal (SIGINT/SIGTERM/SIGHUP) or because every process exited
// with nothing left to restart — then tears everything down and returns.
func (s *Supervisor) Run() error {
	if err := s.Start(); err != nil {
		return err
	}
	s.Wait()
	return nil
}

// Start writes the supervisor pidfile, opens the control socket, and launches
// every process. It returns once everything is started (non-blocking) so tests
// and the foreground runner can drive it. Call Wait to block, or Shutdown to
// tear it down.
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
	}
	if err := s.serveControl(); err != nil {
		return err
	}

	for _, m := range s.procs {
		s.startProc(m)
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
// runs Shutdown.
func (s *Supervisor) Wait() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(sig)
	select {
	case <-sig:
	case <-s.done:
	}
	s.Shutdown()
}

func (s *Supervisor) startProc(m *managed) {
	if s.isStopping() {
		return
	}
	m.mu.Lock()
	if m.active {
		m.mu.Unlock()
		return
	}
	m.active = true
	m.desired = true
	m.restarts = 0
	m.exitCode = nil
	m.status = StatusStarting
	m.mu.Unlock()

	s.mu.Lock()
	s.active++
	s.mu.Unlock()
	s.loops.Add(1)
	go s.runLoop(m)
}

func (s *Supervisor) stopProc(m *managed) {
	m.mu.Lock()
	m.desired = false
	p := m.proc
	m.mu.Unlock()
	if p != nil {
		s.kill(p)
	}
}

func (s *Supervisor) restartProc(m *managed) {
	m.mu.Lock()
	if !m.active {
		m.mu.Unlock()
		s.startProc(m)
		return
	}
	m.gen++
	m.restarts = 0
	m.status = StatusRestarting
	p := m.proc
	m.mu.Unlock()
	if p != nil {
		s.kill(p)
	}
}

func (s *Supervisor) runLoop(m *managed) {
	defer func() {
		m.mu.Lock()
		m.active = false
		m.proc = nil
		m.mu.Unlock()
		s.persist()
		s.decActive()
		s.loops.Done()
	}()

	for {
		m.mu.Lock()
		if !m.desired || s.isStopping() {
			m.status = StatusStopped
			m.mu.Unlock()
			return
		}
		myGen := m.gen
		m.mu.Unlock()

		proc, logFD, err := s.build(m)
		if err != nil {
			m.mu.Lock()
			m.status = StatusCrashed
			m.mu.Unlock()
			logger.Errorf("start %s: %v", m.entry.Name, err)
			return
		}

		now := time.Now()
		m.mu.Lock()
		m.proc = proc
		m.started = &now
		// A process known to listen on a port stays "starting" until watchPorts
		// re-detects it (so a restart isn't "running" while the server is still
		// booting); one that has never listened goes straight to "running".
		if m.expectsPort {
			m.status = StatusStarting
		} else {
			m.status = StatusRunning
		}
		m.mu.Unlock()

		runDone := make(chan struct{})
		go func() { proc.Run(); close(runDone) }()
		awaitStart(proc, runDone)
		s.persist() // pid is now captured
		go s.watchPorts(m, proc, myGen)

		<-runDone
		_ = logFD.Close()
		res := proc.Result()

		m.mu.Lock()
		code := res.ExitCode
		m.exitCode = &code
		m.proc = nil
		m.ports = nil
		genChanged := m.gen != myGen
		desired := m.desired
		restarts := m.restarts
		policy, maxR := m.policy, m.maxRestarts
		m.mu.Unlock()

		ok := res.IsOk()
		switch {
		case s.isStopping() || !desired:
			m.setStatus(StatusStopped)
			return
		case genChanged:
			m.mu.Lock()
			m.restarts = 0
			m.mu.Unlock()
			continue
		case !shouldRestart(policy, ok) || (maxR > 0 && restarts >= maxR):
			m.setStatus(exitStatus(ok))
			return
		}

		m.mu.Lock()
		m.restarts++
		m.status = StatusRestarting
		m.mu.Unlock()
		s.persist()
		select {
		case <-time.After(backoff(restarts + 1)):
		case <-s.done:
		}
	}
}

// awaitStart blocks until the process has a pid (it actually launched) or has
// already exited, so the next persist records a real pid. Capped to keep a
// never-starting command from blocking the run loop forever.
func awaitStart(proc *cexec.Process, runDone <-chan struct{}) {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if proc.Pid() > 0 {
			return
		}
		select {
		case <-runDone:
			return
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// watchPorts polls for the TCP ports the process group led by proc is listening
// on and records the first non-empty set on m, so `proc status` and the
// start/restart progress view can surface "listening on :PORT". It stops at the
// first detection, when the process exits, on shutdown, or after
// portDetectTimeout. gen guards a restart: a stale watcher from an old run won't
// clobber the ports of the current one.
func (s *Supervisor) watchPorts(m *managed, proc *cexec.Process, gen int) {
	deadline := time.Now().Add(portDetectTimeout)
	for time.Now().Before(deadline) {
		if s.isStopping() || !proc.IsRunning() {
			return
		}
		ports, err := utils.ListeningPorts(proc.Pid())
		if err != nil {
			logger.Warnf("detect ports for %s: %v", m.entry.Name, err)
			break
		}
		if len(ports) > 0 {
			m.mu.Lock()
			// m.proc == proc rules out a write landing after the process already
			// exited (runLoop nils m.proc under the same lock); m.gen == gen rules
			// out a stale watcher from a previous run after a restart.
			fresh := m.gen == gen && m.proc == proc
			if fresh {
				m.ports = ports
				m.expectsPort = true
				if m.status == StatusStarting {
					m.status = StatusRunning
				}
			}
			m.mu.Unlock()
			if fresh {
				s.persist()
			}
			return
		}
		select {
		case <-time.After(portPollInterval):
		case <-s.done:
			return
		}
	}
	// The port never came up within the window (or lsof failed). Don't leave a
	// known server wedged in "starting" forever — promote it to "running" so it
	// stays usable; the next start will wait again.
	m.mu.Lock()
	flipped := m.gen == gen && m.proc == proc && m.status == StatusStarting
	if flipped {
		m.status = StatusRunning
	}
	m.mu.Unlock()
	if flipped {
		s.persist()
	}
}

func (s *Supervisor) build(m *managed) (*cexec.Process, *os.File, error) {
	logFD, err := os.OpenFile(m.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log %s: %w", m.logPath, err)
	}
	fmt.Fprintf(logFD, "--- %s start %s ---\n", m.entry.Name, time.Now().Format(time.RFC3339))

	var out io.Writer = logFD
	if s.foreground {
		out = io.MultiWriter(logFD, newPrefixWriter(os.Stdout, &s.stdoutMu, m.entry.Name, s.width, m.colorIdx))
	}
	// The command is run as-is; env vars (from .env / config) are injected via
	// WithEnv and expanded by the shell at runtime. Pre-expanding here would
	// clobber shell-internal variables like $i or arithmetic like $((i+1)).
	proc := clicky.Exec(m.entry.Command).
		WithCwd(s.root).
		WithEnv(m.overlay).
		WithProcessGroup().
		Stream(out, out)
	return proc, logFD, nil
}

// kill terminates a process group gracefully (SIGTERM) then forcefully
// (SIGKILL via KillTree) if it does not exit within stopGrace.
func (s *Supervisor) kill(p *cexec.Process) {
	_ = terminateGroup(p.Pid())
	deadline := time.Now().Add(stopGrace)
	for time.Now().Before(deadline) {
		if !p.IsRunning() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	_ = p.KillTree()
}

// Shutdown stops every process (graceful then forceful), closes the control
// socket, and removes the state/pid files. It is idempotent and does not exit
// the process, so it is safe to call from tests.
func (s *Supervisor) Shutdown() {
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
		m.desired = false
		p := m.proc
		m.mu.Unlock()
		if p == nil {
			continue
		}
		wg.Add(1)
		go func(p *cexec.Process) {
			defer wg.Done()
			s.kill(p)
		}(p)
	}
	wg.Wait()
	s.loops.Wait() // drain run loops so no persist races the cleanup below
	_ = Clean(s.dir)
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
// serialised (the atomic write uses a shared temp file) and skipped once
// shutdown has begun so a late write can't recreate the file Clean just removed.
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
	defer m.mu.Unlock()
	pid := 0
	if m.proc != nil {
		pid = m.proc.Pid()
	}
	return ProcState{
		Name:     m.entry.Name,
		Command:  m.entry.Command,
		PID:      pid,
		Status:   m.status,
		Started:  m.started,
		Restarts: m.restarts,
		ExitCode: m.exitCode,
		LogFile:  m.logPath,
		Ports:    append([]int(nil), m.ports...),
	}
}

func (m *managed) setStatus(status string) {
	m.mu.Lock()
	m.status = status
	m.mu.Unlock()
}

func resolvePolicy(cfg verify.ProcfileConfig, name string) (string, int) {
	policy := cfg.RestartPolicy
	maxR := cfg.MaxRestarts
	if pc, ok := cfg.Processes[name]; ok {
		if pc.RestartPolicy != "" {
			policy = pc.RestartPolicy
		}
		if pc.MaxRestarts != nil {
			maxR = *pc.MaxRestarts
		}
	}
	if policy == "" {
		policy = RestartNo
	}
	return policy, maxR
}

func processEnv(cfg verify.ProcfileConfig, name string) map[string]string {
	if pc, ok := cfg.Processes[name]; ok {
		return pc.Env
	}
	return nil
}

func shouldRestart(policy string, ok bool) bool {
	switch policy {
	case RestartAlways:
		return true
	case RestartOnFailure:
		return !ok
	default:
		return false
	}
}

func exitStatus(ok bool) string {
	if ok {
		return StatusExited
	}
	return StatusCrashed
}

// backoff returns the delay before the nth restart: 500ms doubling up to 30s.
func backoff(n int) time.Duration {
	d := 500 * time.Millisecond
	for i := 1; i < n; i++ {
		d *= 2
		if d >= 30*time.Second {
			return 30 * time.Second
		}
	}
	return d
}

func truncateFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("truncate %s: %w", path, err)
	}
	return f.Close()
}
