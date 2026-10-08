package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// changeNotifier fans a change signal out to every subscribed stream. A
// subscriber's channel holds one pending wake, so a burst of changes wakes a
// slow stream once.
type changeNotifier struct {
	mu   sync.Mutex
	subs map[chan struct{}]struct{}
}

func (n *changeNotifier) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	n.mu.Lock()
	if n.subs == nil {
		n.subs = map[chan struct{}]struct{}{}
	}
	n.subs[ch] = struct{}{}
	n.mu.Unlock()
	return ch, func() {
		n.mu.Lock()
		delete(n.subs, ch)
		n.mu.Unlock()
	}
}

func (n *changeNotifier) Notify() {
	n.mu.Lock()
	defer n.mu.Unlock()
	for ch := range n.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// serveSnapshotStream writes load's result as an SSE data frame on connect,
// on every wake and every interval — only when it differs from the last frame
// sent, else a ping comment. A load failure is sent as an `error` event and
// ends the stream.
func serveSnapshotStream(w http.ResponseWriter, r *http.Request, load func(context.Context) (any, error), interval time.Duration, wake <-chan struct{}) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported by %T", w)
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	var last []byte
	send := func() error {
		value, err := load(r.Context())
		if err != nil {
			payload, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", payload)
			flusher.Flush()
			return err
		}
		frame, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("encode stream frame: %w", err)
		}
		if bytes.Equal(frame, last) {
			fmt.Fprint(w, ": ping\n\n")
		} else {
			fmt.Fprintf(w, "data: %s\n\n", frame)
			last = frame
		}
		flusher.Flush()
		return nil
	}
	if err := send(); err != nil {
		return err
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return nil
		case <-ticker.C:
		case <-wake:
		}
		if err := send(); err != nil {
			return err
		}
	}
}
