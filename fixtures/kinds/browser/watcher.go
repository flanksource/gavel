package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"syscall"
	"time"

	cexec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/fixtures"
)

// Start runs a `background: true` playwright step: the script (e.g.
// e2e/watch-session.ts) runs under node, attached to the shared Chrome, and
// streams NDJSON events on stdout. Start returns once it emitted `ready`; Stop
// (at fixture teardown) collects the verdict.
func (Playwright) Start(ctx context.Context, step *fixtures.StepContext) (fixtures.BackgroundHandle, error) {
	run, err := resolvePlaywright(step, ".ts")
	if err != nil {
		return nil, err
	}
	readyTimeout, stopTimeout := 60*time.Second, 10*time.Second
	if run.cfg.ReadyTimeout != nil {
		readyTimeout = run.cfg.ReadyTimeout.Std()
	}
	if run.cfg.StopTimeout != nil {
		stopTimeout = run.cfg.StopTimeout.Std()
	}
	h := &watcherHandle{step: step, run: run, start: time.Now(), stopTimeout: stopTimeout, ready: make(chan struct{})}
	// Not bound to ctx: the watcher outlives the step that started it.
	h.process = cexec.NewExec("node", run.script).WithCwd(run.cwd).WithExactEnv(run.env).WithProcessGroup().
		Stream(&lineWriter{onLine: h.record}, nil)
	go h.process.Run()

	if err := h.awaitReady(ctx, readyTimeout); err != nil {
		return nil, errors.Join(fmt.Errorf("playwright step %q: %w", step.Spec.ID, err), h.process.KillTree(), run.cleanup())
	}
	return h, nil
}

// watcherHandle is a running background watcher.
type watcherHandle struct {
	step        *fixtures.StepContext
	run         *playwrightRun
	process     *cexec.Process
	start       time.Time
	stopTimeout time.Duration

	mu        sync.Mutex
	events    []map[string]any
	malformed []string
	ready     chan struct{}
	readyOnce sync.Once
}

func (h *watcherHandle) record(line string) {
	if strings.TrimSpace(line) == "" {
		return
	}
	var event map[string]any
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := json.Unmarshal([]byte(line), &event); err != nil || event["type"] == nil {
		h.malformed = append(h.malformed, line)
		return
	}
	// ready is the watcher's lifecycle handshake, not an observed page event, so
	// it never lands in the reported events.
	if event["type"] == "ready" {
		h.readyOnce.Do(func() { close(h.ready) })
		return
	}
	h.events = append(h.events, event)
}

func (h *watcherHandle) eventTypes() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	types := make([]string, 0, len(h.events))
	for _, event := range h.events {
		types = append(types, fmt.Sprint(event["type"]))
	}
	return types
}

func (h *watcherHandle) exited() bool { return h.process.Result().Status != "pending" }

func (h *watcherHandle) awaitReady(ctx context.Context, timeout time.Duration) error {
	deadline := time.After(timeout)
	for {
		select {
		case <-h.ready:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("watcher %s emitted no ready event within %s: %s", h.run.script, timeout, strings.TrimSpace(h.process.GetStderr()))
		case <-time.After(50 * time.Millisecond):
			if h.exited() {
				return fmt.Errorf("watcher %s exited (code %d) before ready: %s", h.run.script, h.process.ExitCode(), strings.TrimSpace(h.process.GetStderr()))
			}
		}
	}
}

// Stop sends SIGTERM, waits for the watcher to flush and exit, and fails the
// step on any alert, console error, page error or failed API response it saw.
func (h *watcherHandle) Stop(context.Context) fixtures.FixtureResult {
	result := newResult(h.step, h.start)
	result.Command, result.CWD = "node "+h.run.script, h.run.cwd
	problems := h.terminate()
	if err := h.run.cleanup(); err != nil {
		problems = append(problems, "remove inline script: "+err.Error())
	}

	h.mu.Lock()
	events := append([]map[string]any(nil), h.events...)
	for _, line := range h.malformed {
		problems = append(problems, "malformed watcher output: "+line)
	}
	h.mu.Unlock()
	for _, event := range events {
		if finding := describeEvent(event); finding != "" {
			problems = append(problems, finding)
		}
	}

	result.Stdout, result.Stderr, result.ExitCode = h.process.GetStdout(), h.process.GetStderr(), h.process.ExitCode()
	result.Metadata = map[string]any{"events": events}
	if len(problems) > 0 {
		result = result.Failf("%s", strings.Join(problems, "\n\n"))
	}
	return finish(result)
}

// terminate ends the watcher; a watcher that already died, ignores SIGTERM or
// exits non-zero is itself a failure (it stopped watching).
func (h *watcherHandle) terminate() []string {
	if h.exited() {
		return []string{fmt.Sprintf("watcher exited before teardown (code %d): %s", h.process.ExitCode(), strings.TrimSpace(h.process.GetStderr()))}
	}
	if err := syscall.Kill(-h.process.Pid(), syscall.SIGTERM); err != nil && !errors.Is(err, syscall.ESRCH) {
		return []string{"signal watcher: " + err.Error()}
	}
	done := make(chan struct{})
	go func() {
		_ = h.process.Wait() // the exit code is checked below
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(h.stopTimeout):
		if err := h.process.KillTree(); err != nil {
			return []string{fmt.Sprintf("watcher ignored SIGTERM for %s and could not be killed: %v", h.stopTimeout, err)}
		}
		<-done
		return []string{fmt.Sprintf("watcher ignored SIGTERM for %s and was killed", h.stopTimeout)}
	}
	if code := h.process.ExitCode(); code != 0 {
		return []string{fmt.Sprintf("watcher exited %d on SIGTERM: %s", code, strings.TrimSpace(h.process.GetStderr()))}
	}
	return nil
}

// describeEvent renders a failing watcher event; ready (and unknown
// informational types) are not failures.
func describeEvent(event map[string]any) string {
	get := func(key string) string {
		if value, ok := event[key]; ok && value != nil {
			return fmt.Sprint(value)
		}
		return ""
	}
	switch get("type") {
	case "alert":
		return strings.TrimSpace(fmt.Sprintf("%s alert on %s:\n%s", get("text"), get("url"), get("details")))
	case "console":
		return fmt.Sprintf("console error on %s: %s", get("url"), get("text"))
	case "pageerror":
		return fmt.Sprintf("page error: %s\n%s", get("message"), get("stack"))
	case "response":
		return fmt.Sprintf("%s %s returned %s", get("method"), get("url"), get("status"))
	default:
		return ""
	}
}

// lineWriter hands each complete line written to it to onLine.
type lineWriter struct {
	mu     sync.Mutex
	buf    []byte
	onLine func(string)
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			return len(p), nil
		}
		line := string(w.buf[:i])
		w.buf = w.buf[i+1:]
		w.onLine(line)
	}
}
