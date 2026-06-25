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
	"github.com/flanksource/gavel/todos/cmux"
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

// todoSessionEvent is the JSON shape pushed to the dashboard's session tab for
// each parsed Claude session-log event.
type todoSessionEvent struct {
	Kind string `json:"kind"`
	Text string `json:"text,omitempty"`
	Tool string `json:"tool,omitempty"`
	// Subagent is the subagent_type for Task/Agent tool calls (e.g. "Explore"),
	// surfaced so the dashboard can filter agent calls by kind independently of
	// the generic "Task" tool name.
	Subagent    string `json:"subagent,omitempty"`
	Action      string `json:"action,omitempty"`
	StopReason  string `json:"stopReason,omitempty"`
	ErrorType   string `json:"errorType,omitempty"`
	ErrorStatus int    `json:"errorStatus,omitempty"`
}

func sessionEventPayload(ev history.SessionEvent) todoSessionEvent {
	out := todoSessionEvent{Kind: string(ev.Kind)}
	switch ev.Kind {
	case history.EventAssistantText, history.EventThinking:
		out.Text = ev.Text
	case history.EventToolUse:
		out.Tool = ev.ToolUse.Tool
		out.Action = history.FormatToolUseSummary(ev.ToolUse.Tool, ev.ToolUse.Input)
		if sub, ok := ev.ToolUse.Input["subagent_type"].(string); ok {
			out.Subagent = sub
		}
	case history.EventTurnEnd:
		out.StopReason = ev.StopReason
	case history.EventError:
		out.Text = ev.Text
		out.StopReason = ev.StopReason
		out.ErrorType = ev.ErrorType
		out.ErrorStatus = ev.ErrorStatus
	}
	return out
}

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
	path, err := cmux.SessionLogPath(dir, sessionID)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	stats, err := cmux.GlobalSessionStats().Get(sessionID, path)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(stats) //nolint:errcheck
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
	if err := cmux.FocusSession(r.Context(), cmux.NewClient(""), dir, agent); err != nil {
		writeTodoError(w, http.StatusBadGateway, err)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"focused": true}) //nolint:errcheck
}

// handleTodoSessionStream follows a TODO's agent session log over SSE. The
// session id is recorded on the issue (session:<id> label) when the run starts,
// so the transcript itself is never stored — the dashboard re-parses it live
// from the log via captain's session parser instead.
func (s *Server) handleTodoSessionStream(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(r.URL.Query().Get("sessionId"))
	if sessionID == "" {
		w.Header().Set("Content-Type", "application/json")
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("sessionId is required"))
		return
	}
	dir := s.resolveTodoDir(strings.TrimSpace(r.URL.Query().Get("dir")))
	path, err := cmux.SessionLogPath(dir, sessionID)
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

// streamSessionLog tails path, parsing each complete line into session events
// and emitting them as SSE `event` frames. It first replays the existing log
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
					events, perr := history.ParseSessionEvents(line)
					if perr != nil {
						continue
					}
					for _, ev := range events {
						progressed = true
						emit("event", sessionEventPayload(ev))
					}
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
