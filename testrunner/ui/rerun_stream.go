package testui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// OutputLine is a single line of process output captured during a rerun.
type OutputLine struct {
	Text   string `json:"text"`
	Stream string `json:"stream"` // "stdout" or "stderr"
}

// RerunOutputBuffer is a thread-safe buffer that captures process output
// during a rerun and serves it to SSE clients.
type RerunOutputBuffer struct {
	mu      sync.Mutex
	lines   []OutputLine
	status  string // "running", "success", "failed", "canceled"
	command string
	updated chan struct{}
}

// NewRerunOutputBuffer creates a buffer ready to capture rerun output.
func NewRerunOutputBuffer(command string) *RerunOutputBuffer {
	return &RerunOutputBuffer{
		status:  "running",
		command: command,
		updated: make(chan struct{}, 1),
	}
}

func (b *RerunOutputBuffer) notify() {
	select {
	case b.updated <- struct{}{}:
	default:
	}
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

type rerunStreamSnapshot struct {
	Lines   []OutputLine `json:"lines"`
	Status  string       `json:"status"`
	Command string       `json:"command"`
}

func (b *RerunOutputBuffer) snapshot(cursor int) (rerunStreamSnapshot, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var lines []OutputLine
	if cursor < len(b.lines) {
		lines = b.lines[cursor:]
	}
	return rerunStreamSnapshot{
		Lines:   lines,
		Status:  b.status,
		Command: b.command,
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
		case <-buf.updated:
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
