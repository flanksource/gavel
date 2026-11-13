package status

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type WatchOptions struct {
	Cache string
	PID   int
	// Interval is how often the status file is polled until the process
	// serves its socket. Defaults to one second.
	Interval time.Duration
}

// Watcher follows one gavel process's Status until the process ends.
type Watcher struct {
	updates chan Status
	err     error
}

// Updates receives each new Status, the first being the current one. It is
// closed when the process ends, the context does, or reading fails (see Err).
func (w *Watcher) Updates() <-chan Status { return w.updates }

// Err is why Updates closed early, once it has closed: nil when the process
// ended or the context did.
func (w *Watcher) Err() error { return w.err }

// Watch follows a process's status file, and its socket's stream once it
// serves one, falling back to the file if the stream breaks.
func Watch(ctx context.Context, options WatchOptions) (*Watcher, error) {
	if options.Interval <= 0 {
		options.Interval = time.Second
	}
	path := Path(options.Cache, options.PID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("watch gavel pid %d: %w", options.PID, err)
	}
	current, err := decode(path, data)
	if err != nil {
		return nil, err
	}
	w := &Watcher{updates: make(chan Status, 1)}
	w.updates <- current
	go func() {
		defer close(w.updates)
		w.err = w.follow(ctx, options, current.Socket, data)
	}()
	return w, nil
}

func (w *Watcher) follow(ctx context.Context, options WatchOptions, socket string, last []byte) error {
	path := Path(options.Cache, options.PID)
	ticker := time.NewTicker(options.Interval)
	defer ticker.Stop()
	for {
		if socket != "" {
			// A dropped connection falls back to the file, which says whether
			// the process is still there; a stream that sent something wrong fails.
			if err := w.stream(ctx, socket); errors.Is(err, errInvalid) {
				return err
			}
			socket = ""
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) || err == nil && !Alive(options.PID) {
			return nil
		}
		if err != nil {
			return err
		}
		if bytes.Equal(data, last) {
			continue
		}
		last = data
		status, err := decode(path, data)
		if err != nil {
			return err
		}
		if !w.send(ctx, status) {
			return nil
		}
		socket = status.Socket
	}
}

func (w *Watcher) send(ctx context.Context, status Status) bool {
	select {
	case <-ctx.Done():
		return false
	case w.updates <- status:
		return true
	}
}

// stream reads the socket's /status/stream until it ends.
func (w *Watcher) stream(ctx context.Context, socket string) error {
	client := http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socket)
		},
	}}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://gavel/status/stream", nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close() //nolint:errcheck // a read-only stream
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("gavel status stream %s: %s", socket, response.Status)
	}
	return readEvents(response.Body, func(event, data string) error {
		if event == "error" {
			return fmt.Errorf("%w: gavel status stream %s: %s", errInvalid, socket, data)
		}
		status, err := decode(socket, []byte(data))
		if err != nil {
			return err
		}
		if !w.send(ctx, status) {
			return ctx.Err()
		}
		return nil
	})
}

// readEvents dispatches each server-sent event's name and data, ignoring comments.
func readEvents(body io.Reader, dispatch func(event, data string) error) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64<<10), 16<<20)
	var event string
	var data []string
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			if len(data) > 0 {
				if err := dispatch(event, strings.Join(data, "\n")); err != nil {
					return err
				}
			}
			event, data = "", nil
		case strings.HasPrefix(line, "event:"):
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
	}
	return scanner.Err()
}

// errInvalid marks what a gavel process published being unreadable, as
// opposed to the process or its connection going away.
var errInvalid = errors.New("invalid gavel status")

func decode(source string, data []byte) (Status, error) {
	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		return Status{}, fmt.Errorf("%w: decode %s: %w", errInvalid, source, err)
	}
	return status, nil
}
