package cmux

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	todopkg "github.com/flanksource/gavel/todos"
)

const (
	// defaultStallTimeout is how long a run may make no progress — neither the
	// session jsonl nor the terminal surface advances — before the watchdog acts.
	defaultStallTimeout = 5 * time.Minute
	// defaultStallNudges is how many times the watchdog re-presses Enter to revive
	// a stalled turn before failing loudly.
	defaultStallNudges = 2
	// defaultStallPollInterval is how often the watchdog samples progress and the
	// surface. It is decoupled from StallTimeout so approval dialogs are detected
	// promptly even while the stall budget is minutes long.
	defaultStallPollInterval = 5 * time.Second
)

// errSessionStalled is returned when the stall watchdog gives up: neither the
// session log nor the terminal surface advanced for StallTimeout, and the
// configured nudges (re-pressing Enter) failed to revive the turn.
var errSessionStalled = errors.New("claude session stalled: no log or surface activity")

// approvalPromptRe matches claude's tool-permission dialog on the terminal
// surface ("Do you want to proceed?", "Do you want to make this edit to …?") —
// the signal that the turn is paused awaiting a human allow/deny.
var approvalPromptRe = regexp.MustCompile(`(?i)\bdo you want to\b`)

// awaitWithStallWatchdog runs awaitSessionCompletion under a dual-signal stall
// watchdog. A watcher goroutine samples the session jsonl (log byte-growth and
// the accumulator's last-activity time) and the cmux surface (readScreen); if
// neither advances for StallTimeout it nudges (re-presses Enter) up to StallNudges
// times, then cancels the wait and reports errSessionStalled. The same watcher
// surfaces claude's tool-permission dialog for human allow/deny. It returns the
// same (logPath, completed, err) as the bare await, except err is errSessionStalled
// when the watchdog gives up.
func (e *CmuxExecutor) awaitWithStallWatchdog(ctx *todopkg.ExecutorContext, ref WorkspaceRef, sessionID, workDir string, timeout time.Duration, resume bool, acc *SessionAccumulator) (string, bool, error) {
	logPath, err := SessionLogPath(workDir, sessionID)
	if err != nil {
		return "", false, err
	}

	// Derive a cancellable context the watchdog can use to unblock the await on a
	// confirmed stall. The shallow copy preserves the logger/transcript/interaction
	// while swapping the embedded Context.
	watchCtx, cancel := context.WithCancel(ctx.Context)
	defer cancel()
	derived := *ctx
	derived.Context = watchCtx

	wd := &stallWatchdog{
		e:         e,
		ref:       ref,
		sessionID: sessionID,
		logPath:   logPath,
		acc:       acc,
		timeout:   e.stallTimeout(),
		maxNudges: e.stallNudges(),
		poll:      e.stallPollInterval(),
	}

	var stalled atomic.Bool
	done := make(chan struct{})
	go func() {
		defer close(done)
		if wd.watch(&derived) {
			stalled.Store(true)
			cancel()
		}
	}()

	path, completed, serr := e.awaitSessionCompletion(&derived, sessionID, workDir, timeout, resume, acc)
	cancel()
	<-done

	if stalled.Load() {
		return path, false, fmt.Errorf("claude session %s made no progress for %s after %d nudge(s): %w", sessionID, wd.timeout, wd.maxNudges, errSessionStalled)
	}
	return path, completed, serr
}

// stallSignal is the sampled progress fingerprint: the accumulator's last
// activity time, the session-log size, and the surface contents. A change in any
// of the three means the run advanced.
type stallSignal struct {
	activity time.Time
	logSize  int64
	screen   string
}

func (s stallSignal) equal(o stallSignal) bool {
	return s.logSize == o.logSize && s.screen == o.screen && s.activity.Equal(o.activity)
}

type stallWatchdog struct {
	e         *CmuxExecutor
	ref       WorkspaceRef
	sessionID string
	logPath   string
	acc       *SessionAccumulator
	timeout   time.Duration
	maxNudges int
	poll      time.Duration

	// approving guards against spawning more than one approval handler for the
	// same on-screen dialog.
	approving atomic.Bool
}

// watch monitors the session for stalls and surfaces tool-permission dialogs. It
// returns true when it has given up — neither signal advanced for StallTimeout
// and all nudges were exhausted — and false when ctx was cancelled (the run
// finished, normally or because the await returned).
func (w *stallWatchdog) watch(ctx *todopkg.ExecutorContext) bool {
	last := w.fingerprint(w.e.readScreen(ctx, w.ref))
	lastProgress := time.Now()
	nudged := 0

	for {
		select {
		case <-ctx.Done():
			return false
		case <-time.After(w.poll):
		}

		screen := w.e.readScreen(ctx, w.ref)
		w.maybeRequestApproval(ctx, screen)

		// While the turn is paused awaiting the user, hold the stall clock — a human
		// thinking about an approval (or an AskUserQuestion) is not a stall — and
		// surface the pause as the "ask" status, since claude renders the approval
		// prompt only on the terminal and no session-log event marks it.
		if w.awaitingHuman(screen) {
			w.markAwaitingHuman()
			lastProgress = time.Now()
			continue
		}

		fp := w.fingerprint(screen)
		if !fp.equal(last) {
			last = fp
			lastProgress = time.Now()
			continue
		}
		if time.Since(lastProgress) < w.timeout {
			continue
		}

		if nudged >= w.maxNudges {
			ctx.Logger.Errorf("cmux: session %s stalled for %s after %d nudge(s); failing", w.sessionID, w.timeout, nudged)
			return true
		}
		nudged++
		ctx.Logger.Warnf("cmux: session %s stalled for %s; nudging (re-pressing Enter) attempt %d/%d", w.sessionID, w.timeout, nudged, w.maxNudges)
		if err := w.e.client.SendKeySurface(ctx, w.ref.String(), w.ref.SurfaceID, "Enter"); err != nil {
			if ctx.Err() != nil {
				return false
			}
			ctx.Logger.Warnf("cmux: session %s nudge failed: %v", w.sessionID, err)
		}
		// Give the nudge time to take before measuring again.
		lastProgress = time.Now()
		last = w.fingerprint(w.e.readScreen(ctx, w.ref))
	}
}

func (w *stallWatchdog) fingerprint(screen string) stallSignal {
	sig := stallSignal{logSize: fileSize(w.logPath), screen: screen}
	if w.acc != nil {
		sig.activity = w.acc.lastActivity()
	}
	return sig
}

// awaitingHuman reports whether the turn is paused on a human: a tool-permission
// dialog on the surface, an approval already in flight, or an ask-tool state.
func (w *stallWatchdog) awaitingHuman(screen string) bool {
	if approvalPromptRe.MatchString(screen) || w.approving.Load() {
		return true
	}
	if w.acc != nil && w.acc.state() == sessionStateAsk {
		return true
	}
	return false
}

// markAwaitingHuman surfaces the paused-on-human condition as the "ask" status on
// the live session, so the dashboard and CLI show "awaiting input" instead of a
// stale "working" while the run waits on an approval. No-op for feedback turns,
// which do not feed the live accumulator.
func (w *stallWatchdog) markAwaitingHuman() {
	if w.acc != nil {
		w.acc.SetState(sessionStateAsk)
	}
}

// maybeRequestApproval surfaces a newly-detected tool-permission dialog for human
// allow/deny via the shared approval registry, spawning the (blocking) handler at
// most once per dialog.
func (w *stallWatchdog) maybeRequestApproval(ctx *todopkg.ExecutorContext, screen string) {
	if !approvalPromptRe.MatchString(screen) {
		return
	}
	if !w.approving.CompareAndSwap(false, true) {
		return
	}
	req := parseApprovalRequest(w.sessionID, screen)
	go func() {
		defer w.approving.Store(false)
		w.e.handleApproval(ctx, w.ref, req)
	}()
}

// handleApproval routes a tool-permission request to the dashboard (via the
// process-wide approval registry) and applies the decision on the surface: Enter
// accepts the highlighted "Yes", Escape cancels (deny). It blocks until the
// dashboard resolves the request or the context is cancelled.
func (e *CmuxExecutor) handleApproval(ctx *todopkg.ExecutorContext, ref WorkspaceRef, req todopkg.ApprovalRequest) {
	ctx.Logger.Infof("cmux: session %s awaiting tool-permission approval: %s", req.SessionID, approvalSummary(req))
	decision, err := todopkg.GlobalApprovals().Await(ctx, req)
	if err != nil {
		ctx.Logger.V(1).Infof("cmux: approval for session %s not resolved: %v", req.SessionID, err)
		return
	}
	if decision.Allow {
		ctx.Logger.Infof("cmux: approval for session %s allowed; accepting on surface", req.SessionID)
		if err := e.client.SendKeySurface(ctx, ref.String(), ref.SurfaceID, "Enter"); err != nil {
			ctx.Logger.Warnf("cmux: failed to send approval accept key: %v", err)
		}
		return
	}
	ctx.Logger.Infof("cmux: approval for session %s denied; cancelling on surface", req.SessionID)
	if err := e.client.SendKeySurface(ctx, ref.String(), ref.SurfaceID, "Escape"); err != nil {
		ctx.Logger.Warnf("cmux: failed to send approval deny key: %v", err)
	}
}

// parseApprovalRequest builds an ApprovalRequest from the dialog text on the
// surface. The tool and input are best-effort — the cmux surface carries only
// rendered text — but enough for the dashboard to show what is being approved.
func parseApprovalRequest(sessionID, screen string) todopkg.ApprovalRequest {
	line := approvalPromptLine(screen)
	return todopkg.ApprovalRequest{
		SessionID: sessionID,
		Tool:      approvalTool(line),
		Input:     map[string]any{"prompt": line},
	}
}

// approvalPromptLine extracts the "Do you want to …" line from the dialog,
// stripping box-drawing borders, or a generic fallback when it cannot be found.
func approvalPromptLine(screen string) string {
	for _, raw := range strings.Split(screen, "\n") {
		line := strings.TrimSpace(strings.Trim(strings.TrimSpace(raw), "│|╮╭╯╰"))
		if approvalPromptRe.MatchString(line) {
			return line
		}
	}
	return "tool permission requested"
}

// approvalTool maps the dialog's prompt line to a coarse tool name for display.
func approvalTool(line string) string {
	l := strings.ToLower(line)
	switch {
	case strings.Contains(l, "make this edit"):
		return "Edit"
	case strings.Contains(l, "create"):
		return "Write"
	case strings.Contains(l, "command") || strings.Contains(l, "run"):
		return "Bash"
	default:
		return "tool"
	}
}

func approvalSummary(req todopkg.ApprovalRequest) string {
	if p, ok := req.Input["prompt"].(string); ok && p != "" {
		return p
	}
	return req.Tool + " permission requested"
}

func (e *CmuxExecutor) stallTimeout() time.Duration {
	if e.config.StallTimeout > 0 {
		return e.config.StallTimeout
	}
	return defaultStallTimeout
}

func (e *CmuxExecutor) stallNudges() int {
	if e.config.StallNudges > 0 {
		return e.config.StallNudges
	}
	return defaultStallNudges
}

func (e *CmuxExecutor) stallPollInterval() time.Duration {
	if e.config.StallPollInterval > 0 {
		return e.config.StallPollInterval
	}
	return defaultStallPollInterval
}
