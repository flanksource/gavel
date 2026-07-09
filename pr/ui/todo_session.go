package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/ai/history"
	cmuxprov "github.com/flanksource/captain/pkg/ai/provider/cmux"
	"github.com/flanksource/gavel/todos"
)

const (
	// sessionStreamPoll is how often the tailer re-reads the session log for new
	// lines and emits a keep-alive when there is nothing new.
	sessionStreamPoll = 500 * time.Millisecond
	// sessionLogAppearTimeout bounds how long a just-started run is given to
	// create its session log before the stream reports it missing.
	sessionLogAppearTimeout = 60 * time.Second
)

// errSessionLogMissing signals the session log never appeared within the
// appear timeout (e.g. a stale/unknown session id, or a run that never started).
var errSessionLogMissing = errors.New("session log did not appear")

// handleTodoSessionStats returns the rolled-up stats for a TODO's agent session
// — agent/model/effort, elapsed time, token usage and derived cost. Live runs are
// served from the in-memory cache the cmux tailer feeds; sessions no tailer is
// watching are read (and cached) from the on-disk log. A session that never
// produced a log is reported as found=false, not an error, so the dashboard
// simply hides the timer.
func (s *Server) handleTodoSessionStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := strings.TrimSpace(r.URL.Query().Get("sessionId"))
	if sessionID == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("sessionId is required"))
		return
	}
	dir := s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir")))
	path, err := cmuxprov.SessionLogPath(dir, sessionID)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	stats, err := cmuxprov.GlobalSessionStats().Get(sessionID, path)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	resp := todoSessionStatsResponse{SessionStats: stats}
	// A pending tool-permission request overrides the derived state so the
	// dashboard can render the "Needs approval" affordance and its Allow/Deny
	// buttons regardless of which driver produced it.
	if req, ok := todos.GlobalApprovals().Pending(sessionID); ok {
		resp.State = "approval"
		pending := req
		resp.Approval = &pending
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// todoSessionStatsResponse is the session-stats payload plus any pending
// tool-permission request awaiting the user's Allow/Deny.
type todoSessionStatsResponse struct {
	cmuxprov.SessionStats
	Approval *todos.ApprovalRequest `json:"approval,omitempty"`
}

// handleTodoSessionApprove resolves a pending tool-permission request for a
// session — the dashboard's Allow/Deny buttons POST here, which unblocks the
// driver awaiting the decision (see todos.ApprovalRegistry).
func (s *Server) handleTodoSessionApprove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload struct {
		SessionID    string         `json:"sessionId"`
		Allow        bool           `json:"allow"`
		Message      string         `json:"message,omitempty"`
		UpdatedInput map[string]any `json:"updatedInput,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	sessionID := strings.TrimSpace(payload.SessionID)
	if sessionID == "" {
		sessionID = strings.TrimSpace(r.URL.Query().Get("sessionId"))
	}
	if sessionID == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("sessionId is required"))
		return
	}
	if err := todos.GlobalApprovals().Resolve(sessionID, todos.ApprovalDecision{
		Allow:        payload.Allow,
		Message:      payload.Message,
		UpdatedInput: payload.UpdatedInput,
	}); err != nil {
		writeTodoError(w, http.StatusConflict, err)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"resolved": true, "allow": payload.Allow}) //nolint:errcheck
}

// handleTodoSessionFocus switches cmux to the workspace running a TODO's agent
// session, so the dashboard's "focus" control brings the live terminal to the
// front. The workspace is identified by the run's working directory and agent
// (claude/codex), matching how the cmux executor names it. A closed terminal or
// a stopped cmux yields a 4xx with the reason rather than a silent no-op.
func (s *Server) handleTodoSessionFocus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dir := s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir")))
	agent := strings.TrimSpace(r.URL.Query().Get("agent"))
	if agent == "" {
		agent = "claude"
	}
	if err := cmuxprov.FocusSession(r.Context(), cmuxprov.NewClient(""), dir, agent); err != nil {
		writeTodoError(w, http.StatusBadGateway, err)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"focused": true}) //nolint:errcheck
}

// todoSessionCmuxResponse locates the cmux terminal a session maps to. It is
// best-effort: a stopped cmux or a closed terminal reports found=false with the
// reason rather than an error, so the dashboard's cmux control still renders (and
// can offer Resume) instead of the whole request failing.
type todoSessionCmuxResponse struct {
	Found     bool   `json:"found"`
	Workspace string `json:"workspace,omitempty"`
	Surface   string `json:"surface,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// handleTodoSessionCmux resolves the cmux workspace and surface a TODO's agent
// session lives on, so the dashboard can label the "Focus / Resume in cmux"
// control with the concrete terminal it targets (e.g. "workspace:2 surface:1").
// Resolution reuses the same workspace naming the cmux executor uses, so no cmux
// reference has to be tracked on the issue.
func (s *Server) handleTodoSessionCmux(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dir := s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir")))
	agent := strings.TrimSpace(r.URL.Query().Get("agent"))
	if agent == "" {
		agent = "claude"
	}
	workspace, surface, err := resolveCmuxSurface(r.Context(), dir, agent)
	resp := todoSessionCmuxResponse{Found: err == nil}
	if err != nil {
		resp.Reason = err.Error()
	} else {
		resp.Workspace = workspace
		resp.Surface = surface
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// resolveCmuxSurface finds the cmux workspace running the agent session for
// workDir — named the same way the cmux executor names it — and the surface its
// terminal currently shows. The surface read is best-effort: a workspace whose
// tree can't be read still yields the workspace reference.
func resolveCmuxSurface(ctx context.Context, workDir, agent string) (workspace, surface string, err error) {
	client := cmuxprov.NewClient("")
	if err := client.Available(ctx); err != nil {
		return "", "", err
	}
	name := cmuxprov.AgentWorkspaceName(workDir, agent)
	ref, found, err := client.FindWorkspace(ctx, name, workDir)
	if err != nil {
		return "", "", err
	}
	if !found {
		return "", "", fmt.Errorf("no cmux workspace %q for %s; the session terminal may have been closed", name, workDir)
	}
	surface, _ = client.ResolveSurface(ctx, ref.String())
	return ref.String(), surface, nil
}

// handleTodoSessionStream follows a TODO's agent session log over SSE. The
// session id is recorded on the issue (session:<id> label) when the run starts,
// so the transcript itself is never stored — the dashboard streams the raw
// captain session entries and renders them with clicky-ui's SessionViewer.
func (s *Server) handleTodoSessionStream(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(r.URL.Query().Get("sessionId"))
	if sessionID == "" {
		w.Header().Set("Content-Type", "application/json")
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("sessionId is required"))
		return
	}
	dir := s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir")))
	path, err := cmuxprov.SessionLogPath(dir, sessionID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	streamSessionLog(w, r, flusher, path)
}

// streamSessionLog tails path, parsing each complete line into a captain
// SessionEntry and emitting the conversational ones as SSE `entry` frames (the
// schema clicky-ui's SessionViewer consumes). It first replays the existing log
// (so reopening the tab shows full history) then follows appended lines until
// the client disconnects. Unlike the executor's tailer it does not stop at
// end_turn — a resumed run keeps streaming into the same log.
func streamSessionLog(w http.ResponseWriter, r *http.Request, flusher http.Flusher, path string) {
	emit := func(event string, data any) {
		b, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
		flusher.Flush()
	}

	f, err := openSessionLog(r.Context(), path)
	if err != nil {
		if errors.Is(err, errSessionLogMissing) {
			emit("error", map[string]string{"error": "no session activity yet"})
		}
		return
	}
	defer func() { _ = f.Close() }()

	var pending []byte
	buf := make([]byte, 32*1024)
	for {
		progressed := false
		for {
			n, rerr := f.Read(buf)
			if n > 0 {
				pending = append(pending, buf[:n]...)
				for {
					i := bytes.IndexByte(pending, '\n')
					if i < 0 {
						break
					}
					line := pending[:i]
					pending = pending[i+1:]
					// Emit the raw captain entry for the SessionViewer, but only for
					// conversational (assistant) lines — Events() is empty for user
					// tool-results and bookkeeping, which the viewer would drop anyway.
					var entry history.SessionEntry
					if json.Unmarshal(line, &entry) != nil || len(entry.Events()) == 0 {
						continue
					}
					progressed = true
					emit("entry", entry)
				}
			}
			if rerr == io.EOF {
				break
			}
			if rerr != nil {
				emit("error", map[string]string{"error": rerr.Error()})
				return
			}
		}
		if !progressed {
			// Keep-alive comment frame: holds the socket open without firing a
			// client-side message handler.
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
		select {
		case <-r.Context().Done():
			return
		case <-time.After(sessionStreamPoll):
		}
	}
}

// openSessionLog waits for the session log to exist, bounded by the appear
// timeout and the request context. A growing file returns plain io.EOF at the
// tail, so callers can keep reading for appended lines.
func openSessionLog(ctx context.Context, path string) (*os.File, error) {
	deadline := time.Now().Add(sessionLogAppearTimeout)
	for {
		f, err := os.Open(path)
		if err == nil {
			return f, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
		if time.Now().After(deadline) {
			return nil, errSessionLogMissing
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(sessionStreamPoll):
		}
	}
}
