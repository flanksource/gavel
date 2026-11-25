package testui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/flanksource/clicky/sse"
	"github.com/flanksource/gavel/utils"
)

// OutputLine is a single line of process output captured during a rerun.
type OutputLine struct {
	Text   string `json:"text"`
	Stream string `json:"stream"` // "stdout" or "stderr"
}

// RerunCommand is a process a rerun ran: Line is argv quoted for a shell, and
// Shell the line that reruns it from any directory (cd to Cwd first).
type RerunCommand struct {
	Cwd   string   `json:"cwd,omitempty"`
	Argv  []string `json:"argv"`
	Line  string   `json:"line"`
	Shell string   `json:"shell"`
}

func newRerunCommand(cwd string, argv []string) RerunCommand {
	line := utils.ShellJoin(argv)
	shell := line
	if cwd != "" {
		shell = "cd " + utils.ShellQuote(cwd) + " && " + line
	}
	return RerunCommand{Cwd: cwd, Argv: slices.Clone(argv), Line: line, Shell: shell}
}

// RerunOutputBuffer is a thread-safe buffer that captures process output
// during a rerun and serves it to SSE clients.
type RerunOutputBuffer struct {
	mu       sync.Mutex
	lines    []OutputLine
	commands []RerunCommand
	status   string // "running", "success", "failed", "canceled"
	command  string
	updated  sse.Notifier
}

// NewRerunOutputBuffer creates a buffer ready to capture rerun output.
func NewRerunOutputBuffer(command string) *RerunOutputBuffer {
	return &RerunOutputBuffer{
		status:  "running",
		command: command,
	}
}

func (b *RerunOutputBuffer) notify() {
	b.updated.Notify()
}

// StdoutWriter returns an io.Writer that appends to the buffer as stdout.
func (b *RerunOutputBuffer) StdoutWriter() *streamWriter {
	return &streamWriter{buf: b, stream: "stdout"}
}

// StderrWriter returns an io.Writer that appends to the buffer as stderr.
func (b *RerunOutputBuffer) StderrWriter() *streamWriter {
	return &streamWriter{buf: b, stream: "stderr"}
}

type streamWriter struct {
	buf    *RerunOutputBuffer
	stream string
}

// RecordCommand makes the writer a utils.CommandRecorder, keeping each command
// the rerun runs into it.
func (w *streamWriter) RecordCommand(cwd string, argv []string) {
	w.buf.mu.Lock()
	w.buf.commands = append(w.buf.commands, newRerunCommand(cwd, argv))
	w.buf.mu.Unlock()
	w.buf.notify()
}

func (w *streamWriter) Write(p []byte) (int, error) {
	w.buf.mu.Lock()
	w.buf.lines = append(w.buf.lines, OutputLine{Text: string(p), Stream: w.stream})
	w.buf.mu.Unlock()
	w.buf.notify()
	return len(p), nil
}

// Finish marks the buffer as done with the given status.
func (b *RerunOutputBuffer) Finish(success bool) {
	b.mu.Lock()
	if b.status == "canceled" {
		b.mu.Unlock()
		b.notify()
		return
	}
	if success {
		b.status = "success"
	} else {
		b.status = "failed"
	}
	b.mu.Unlock()
	b.notify()
}

// Cancel marks the buffer as canceled by the user.
func (b *RerunOutputBuffer) Cancel() {
	b.mu.Lock()
	b.status = "canceled"
	b.mu.Unlock()
	b.notify()
}

// rerunStreamSnapshot is one SSE frame: Lines are the output written since the
// last frame, Commands every command the rerun has run so far.
type rerunStreamSnapshot struct {
	Lines    []OutputLine   `json:"lines"`
	Commands []RerunCommand `json:"commands"`
	Status   string         `json:"status"`
	Command  string         `json:"command"`
}

func (b *RerunOutputBuffer) snapshot(cursor int) (rerunStreamSnapshot, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var lines []OutputLine
	if cursor < len(b.lines) {
		lines = b.lines[cursor:]
	}
	return rerunStreamSnapshot{
		Lines:    lines,
		Commands: slices.Clone(b.commands),
		Status:   b.status,
		Command:  b.command,
	}, len(b.lines)
}

func (s *Server) handleRerunStream(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	buf := s.rerunOutput
	s.mu.RUnlock()

	if buf == nil {
		http.Error(w, "no rerun in progress", http.StatusNotFound)
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

	wake, unsubscribe := buf.updated.Subscribe()
	defer unsubscribe()

	cursor := 0
	snap, cursor := buf.snapshot(cursor)
	if b, err := json.Marshal(snap); err == nil {
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-wake:
		case <-ticker.C:
		}

		snap, cursor = buf.snapshot(cursor)
		if len(snap.Lines) > 0 || snap.Status != "running" {
			if b, err := json.Marshal(snap); err == nil {
				fmt.Fprintf(w, "data: %s\n\n", b)
				flusher.Flush()
			}
		}

		if snap.Status != "running" {
			fmt.Fprintf(w, "event: done\ndata: {}\n\n")
			flusher.Flush()
			return
		}
	}
}
