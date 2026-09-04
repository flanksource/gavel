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
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/session"
	"github.com/google/uuid"
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
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	if store, ok := todoCaptainSessionStore(r.Context(), dir); ok {
		resolved, err := resolveCaptainSession(r.Context(), store, sessionID)
		if err != nil {
			writeTodoSessionStreamError(w, flusher, err)
			return
		}
		if resolved.known {
			streamCaptainSession(w, r, store, sessionID, resolved)
			return
		}
	}
	path, err := cmuxprov.SessionLogPath(dir, sessionID)
	if err != nil {
		writeTodoSessionStreamError(w, flusher, err)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	streamSessionLog(w, r, flusher, path)
}

func writeTodoSessionStreamError(w http.ResponseWriter, flusher http.Flusher, err error) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	payload, marshalErr := json.Marshal(map[string]string{"error": err.Error()})
	if marshalErr != nil {
		panic(fmt.Errorf("marshal session stream error: %w", marshalErr))
	}
	fmt.Fprintf(w, "event: error\ndata: %s\n\n", payload)
	flusher.Flush()
}

// streamCaptainSession replays and follows the monitor-owned unified message
// projection. It emits a row again when a later ingest enriches the same
// message (for example, when tool output arrives); the browser deduplicates by
// message ID and replaces the earlier version.
func streamCaptainSession(w http.ResponseWriter, r *http.Request, store captainSessionStore, sessionID string, resolved captainSessionResolution) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	emit := func(event string, data any) {
		b, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
		flusher.Flush()
	}
	for _, diagnostic := range resolved.diagnostics {
		emit("warning", diagnostic)
	}
	deadline := time.Now().Add(sessionLogAppearTimeout)
	seen := map[uuid.UUID]string{}
	owners := map[uuid.UUID]*captaindb.Session{}
	for {
		if resolved.transcript == nil {
			next, err := resolveCaptainSession(r.Context(), store, sessionID)
			if err != nil {
				emit("error", map[string]string{"error": err.Error()})
				return
			}
			resolved = next
			if resolved.transcript == nil && time.Now().After(deadline) {
				emit("error", map[string]string{"error": "no monitored session activity yet"})
				return
			}
		}
		if resolved.transcript != nil {
			var rows []captaindb.TranscriptMessage
			var err error
			if threadStore, ok := store.(captainThreadSessionStore); ok {
				rows, err = threadStore.ListThreadTranscriptMessages(r.Context(), resolved.transcript.ID)
			} else {
				rows, err = store.ListTranscriptMessages(r.Context(), captaindb.TranscriptPage{SessionID: resolved.transcript.ID})
			}
			if err != nil {
				emit("error", map[string]string{"error": err.Error()})
				return
			}
			for _, row := range rows {
				owner := owners[row.SessionID]
				if owner == nil && row.SessionID == resolved.transcript.ID {
					owner = resolved.transcript
					owners[row.SessionID] = owner
				}
				if owner == nil {
					owner, err = store.GetSessionByIdentity(r.Context(), row.SessionID.String(), "", "", "")
					if err != nil {
						emit("error", map[string]string{"error": err.Error()})
						return
					}
					owners[row.SessionID] = owner
				}
				message, fingerprint, err := captainTranscriptMessage(row, owner)
				if err != nil {
					continue
				}
				if seen[row.ID] == fingerprint {
					continue
				}
				seen[row.ID] = fingerprint
				emit("entry", message)
			}
		}
		if len(seen) == 0 {
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

func captainTranscriptMessage(row captaindb.TranscriptMessage, owner *captaindb.Session) (session.Message, string, error) {
	var parts []session.Part
	if err := json.Unmarshal(row.Parts, &parts); err != nil {
		return session.Message{}, "", err
	}
	message := session.Message{
		ID: row.ID.String(), Role: row.Role, Parts: parts,
		Provenance: &session.Provenance{CWD: owner.CWD, Source: owner.Source, SessionID: owner.ProviderSessionID, AgentID: owner.ID.String()},
		AgentID:    owner.ID.String(),
	}
	if row.TurnID != nil {
		message.TurnID = row.TurnID.String()
	}
	if row.OccurredAt != nil {
		message.Provenance.Timestamp = row.OccurredAt
	}
	if row.Model != nil {
		message.Provenance.Model = *row.Model
	}
	if row.SourceLine != nil {
		message.SourceLine = *row.SourceLine
	}
	fingerprintBytes, err := json.Marshal(message)
	if err != nil {
		return session.Message{}, "", err
	}
	return message, string(fingerprintBytes), nil
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
